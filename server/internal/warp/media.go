package warp

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/abema/go-mp4"
	"github.com/kixelated/invoker"
	"github.com/zencoder/go-dash/v3/mpd"
)

// This is a demo; you should actually fetch media from a live backend.
// It's just much easier to read from disk and "fake" being live.
type Media struct {
	base        fs.FS
	inits       map[string]*MediaInit
	video       []*mpd.Representation
	audio       []*mpd.Representation
	isStreaming bool
}

func NewMedia(playlistPath string, isStreaming bool) (m *Media, err error) {
	m = new(Media)
	m.isStreaming = isStreaming

	// Create a fs.FS out of the folder holding the playlist
	m.base = os.DirFS(filepath.Dir(playlistPath))

	// Read the playlist file
	playlist, err := mpd.ReadFromFile(playlistPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open playlist: %w", err)
	}

	if len(playlist.Periods) > 1 {
		return nil, fmt.Errorf("multiple periods not supported")
	}

	period := playlist.Periods[0]

	for _, adaption := range period.AdaptationSets {
		representation := adaption.Representations[0]

		if representation.MimeType == nil {
			return nil, fmt.Errorf("missing representation mime type")
		}

		if representation.Bandwidth == nil {
			return nil, fmt.Errorf("missing representation bandwidth")
		}

		switch *representation.MimeType {
		case "video/mp4":
			m.video = append(m.video, representation)
		case "audio/mp4":
			m.audio = append(m.audio, representation)
		}
	}

	if len(m.video) == 0 {
		return nil, fmt.Errorf("no video representation found")
	}

	if len(m.audio) == 0 {
		return nil, fmt.Errorf("no audio representation found")
	}

	m.inits = make(map[string]*MediaInit)

	var reps []*mpd.Representation
	reps = append(reps, m.audio...)
	reps = append(reps, m.video...)

	for _, rep := range reps {
		path := *rep.SegmentTemplate.Initialization

		// TODO Support the full template engine
		path = strings.ReplaceAll(path, "$RepresentationID$", *rep.ID)

		f, err := fs.ReadFile(m.base, path)
		if err != nil {
			return nil, fmt.Errorf("failed to read init file: %w", err)
		}

		init, err := newMediaInit(*rep.ID, f)
		if err != nil {
			return nil, fmt.Errorf("failed to create init segment: %w", err)
		}

		m.inits[*rep.ID] = init
	}

	return m, nil
}

func (m *Media) Start(bitrate func() uint64) (inits map[string]*MediaInit, audio *MediaStream, video *MediaStream, err error) {
	audio, err = newMediaStream(m, m.audio, bitrate)
	if err != nil {
		return nil, nil, nil, err
	}

	video, err = newMediaStream(m, m.video, bitrate)
	if err != nil {
		return nil, nil, nil, err
	}

	return m.inits, audio, video, nil
}

func findLatestSequence(dir fs.FS) (int, error) {
	matches, _ := fs.Glob(dir, "chunk-stream*.m4s")
	largest := -1
	for _, match := range matches {
		id, err := strconv.Atoi(strings.TrimSuffix(strings.TrimLeft(strings.Split(match, "-")[2], "0"), ".m4s"))
		if err == nil && id > largest {
			largest = id
		}
	}

	if largest == -1 {
		return -1, errors.New("couldn't find latest sequence")
	}

	return largest, nil
}

var DEFAULT_START_TIME = time.Time{}

type MediaStream struct {
	Media *Media

	start        time.Time
	streamOffset time.Duration
	reps         []*mpd.Representation
	sequence     int
	bitrate      func() uint64 // returns the current estimated bitrate
}

func newMediaStream(m *Media, reps []*mpd.Representation, bitrate func() uint64) (ms *MediaStream, err error) {
	ms = new(MediaStream)
	ms.Media = m
	ms.reps = reps
	ms.start = DEFAULT_START_TIME
	ms.bitrate = bitrate
	ms.streamOffset = -1
	if m.isStreaming {
		latestSequence, err := findLatestSequence(m.base)
		if err != nil {
			fmt.Println("WARN: Cannot find latest sequence, defaulting to 1")
		} else {
			ms.sequence = max(latestSequence+1, 1)
		}
	}
	return ms, nil
}

func (ms *MediaStream) chooseRepresentation(preferredId string) (choice *mpd.Representation) {
	bitrate := ms.bitrate()

	// Loop over the renditions and pick the highest bitrate we can support
	for _, r := range ms.reps {
		if *r.ID == preferredId {
			choice = r
			break
		} else if uint64(*r.Bandwidth) <= bitrate && (choice == nil || *r.Bandwidth > *choice.Bandwidth) {
			choice = r
		}
	}

	if choice != nil {
		return choice
	}

	// We can't support any of the bitrates, so find the lowest one.
	for _, r := range ms.reps {
		if choice == nil || *r.Bandwidth < *choice.Bandwidth {
			choice = r
		}
	}

	return choice
}

// Returns the next segment in the stream
func (ms *MediaStream) Next(ctx context.Context, session *Session, timeOffset time.Duration) (segment *MediaSegment, err error) {
	rep := ms.chooseRepresentation(session.prefs["resolution"])

	if rep.SegmentTemplate == nil {
		return nil, fmt.Errorf("missing segment template")
	}

	if rep.SegmentTemplate.Media == nil {
		return nil, fmt.Errorf("no media template")
	}

	if rep.SegmentTemplate.StartNumber == nil {
		return nil, fmt.Errorf("missing start number")
	}

	path := *rep.SegmentTemplate.Media
	sequence := ms.sequence + int(*rep.SegmentTemplate.StartNumber)

	// TODO Support the full template engine
	path = strings.ReplaceAll(path, "$RepresentationID$", *rep.ID)
	path = strings.ReplaceAll(path, "$Number%05d$", fmt.Sprintf("%05d", sequence)) // TODO TODO

	// This loop is there for actual live stream purposes, if there's no more
	// segment that shows up after 1 minute, assume the streamer has either ended
	// the stream, or got disconnected
	start := time.Now()
	var f fs.File
	for {
		// Try openning the file
		mediaFile, err := ms.Media.base.Open(path)
		if err == nil {
			f = mediaFile
			break
		}

		var errorMsg error
		if errors.Is(err, os.ErrNotExist) && ms.sequence != 0 {
			if session.server.isStreaming {
				errorMsg = fmt.Errorf("timed out, disconnected?")
			} else {
				errorMsg = fmt.Errorf("EOF")
			}
		} else {
			errorMsg = fmt.Errorf("failed to open segment file: %w", err)
		}

		if time.Since(start) >= 1*time.Minute || !session.server.isStreaming {
			return nil, errorMsg
		}
	}

	duration := time.Duration(*rep.SegmentTemplate.Duration) / time.Nanosecond
	timestamp := time.Duration(ms.sequence)*duration + timeOffset

	init := ms.Media.inits[*rep.ID]

	segment, err = newMediaSegment(ms, init, f, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to create segment: %w", err)
	}

	ms.sequence += 1

	return segment, nil
}

type MediaInit struct {
	ID        string
	Raw       []byte
	Timescale int
}

func newMediaInit(id string, raw []byte) (mi *MediaInit, err error) {
	mi = new(MediaInit)
	mi.ID = id
	mi.Raw = raw

	err = mi.parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse init segment: %w", err)
	}

	return mi, nil
}

// Parse through the init segment, literally just to populate the timescale
func (mi *MediaInit) parse() (err error) {
	r := bytes.NewReader(mi.Raw)

	_, err = mp4.ReadBoxStructure(r, func(h *mp4.ReadHandle) (interface{}, error) {
		if !h.BoxInfo.IsSupportedType() {
			return nil, nil
		}

		payload, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}

		switch box := payload.(type) {
		case *mp4.Mdhd: // Media Header; moov -> trak -> mdia > mdhd
			if mi.Timescale != 0 {
				// verify only one track
				return nil, fmt.Errorf("multiple mdhd atoms")
			}

			mi.Timescale = int(box.Timescale)
		}

		// Expands children
		return h.Expand()
	})

	if err != nil {
		return fmt.Errorf("failed to parse MP4 file: %w", err)
	}

	return nil
}

type MediaSegment struct {
	Stream *MediaStream
	Init   *MediaInit

	file      fs.File
	timestamp time.Duration
}

func newMediaSegment(s *MediaStream, init *MediaInit, file fs.File, timestamp time.Duration) (ms *MediaSegment, err error) {
	ms = new(MediaSegment)
	ms.Stream = s
	ms.Init = init

	ms.file = file
	ms.timestamp = timestamp

	return ms, nil
}

// Return the next atom, sleeping based on the PTS to simulate a live stream
func (ms *MediaSegment) Read(ctx context.Context) (chunk []byte, err error) {
	// Read the next top-level box
	var header [8]byte

	_, err = io.ReadFull(ms.file, header[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	size := int(binary.BigEndian.Uint32(header[0:4]))
	if size < 8 {
		return nil, fmt.Errorf("box is too small")
	}

	buf := make([]byte, size)
	n := copy(buf, header[:])

	_, err = io.ReadFull(ms.file, buf[n:])
	if err != nil {
		return nil, fmt.Errorf("failed to read atom: %w", err)
	}

	sample, err := ms.parseAtom(ctx, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse atom: %w", err)
	}

	if sample != nil {
		// Only running once, set initial stream offset
		if ms.Stream.Media.isStreaming && ms.Stream.streamOffset == -1 && ms.Stream.start == DEFAULT_START_TIME {
			ms.Stream.streamOffset = sample.Timestamp
			ms.Stream.start = time.Now()
		}

		offset := 0 * time.Nanosecond
		if ms.Stream.Media.isStreaming {
			offset = ms.Stream.streamOffset
		}

		// Simulate a live stream by sleeping before we write this sample.
		// Figure out how much time has elapsed since the start
		elapsed := time.Since(ms.Stream.start)
		delay := (sample.Timestamp - elapsed - offset)

		if delay > 0 {
			// Sleep until we're supposed to see these samples
			err = invoker.Sleep(delay)(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	return buf, nil
}

// Parse through the MP4 atom, returning infomation about the next fragmented sample
func (ms *MediaSegment) parseAtom(ctx context.Context, buf []byte) (sample *mediaSample, err error) {
	r := bytes.NewReader(buf)

	_, err = mp4.ReadBoxStructure(r, func(h *mp4.ReadHandle) (interface{}, error) {
		if !h.BoxInfo.IsSupportedType() {
			return nil, nil
		}

		payload, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}

		switch box := payload.(type) {
		case *mp4.Moof:
			sample = new(mediaSample)
		case *mp4.Tfdt: // Track Fragment Decode Timestamp; moof -> traf -> tfdt
			// TODO This box isn't required
			// TODO we want the last PTS if there are multiple samples
			var dts time.Duration
			if box.FullBox.Version == 0 {
				dts = time.Duration(box.BaseMediaDecodeTimeV0)
			} else {
				dts = time.Duration(box.BaseMediaDecodeTimeV1)
			}

			if ms.Init.Timescale == 0 {
				return nil, fmt.Errorf("missing timescale")
			}

			// Convert to seconds
			// TODO What about PTS?
			sample.Timestamp = dts * time.Second / time.Duration(ms.Init.Timescale)
		}

		// Expands children
		return h.Expand()
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse MP4 file: %w", err)
	}

	return sample, nil
}

func (ms *MediaSegment) Close() (err error) {
	return ms.file.Close()
}

type mediaSample struct {
	Timestamp time.Duration // The timestamp of the first sample
}
