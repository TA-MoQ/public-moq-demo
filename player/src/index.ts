import { Player } from "./player";
import * as Plotly from "plotly.js-dist";
import { estimator } from "./estimator";

// This is so ghetto but I'm too lazy to improve it right now
const vidRef = document.getElementById("vid") as HTMLVideoElement;
const startRef = document.getElementById("start") as HTMLButtonElement;
const liveRef = document.getElementById("live") as HTMLButtonElement;
const throttleRef = document.getElementById("throttle") as HTMLDivElement;
const throttleDDL = document.getElementById("throttles") as HTMLSelectElement;
const statsRef = document.getElementById("stats") as HTMLDivElement;
const resolutionsRef = document.getElementById(
  "resolutions"
) as HTMLSelectElement;
const activeBWTestRef = document.getElementById("active_bw_test");
const continueStreamingRef = document.getElementById("continue_streaming");
const logContentRef = document.querySelector(
  "#log_content"
) as HTMLTextAreaElement;
const toggleLogRef = document.querySelector("#toggle_log") as HTMLAnchorElement;

// const datagramRef = document.getElementById("test_datagram") as HTMLButtonElement;
// const stremRef = document.getElementById("test_stream") as HTMLButtonElement;
// const datagramRef2 = document.getElementById("test_datagram2") as HTMLButtonElement;
// const stremRef2 = document.getElementById("test_stream2") as HTMLButtonElement;

const categoryRef = document.getElementById("category") as HTMLSelectElement;

const params = new URLSearchParams(window.location.search);
window.estimator = estimator;

if (process.env.SERVER_URL) {
  console.log("Setting server url to %s", process.env.SERVER_URL);
  window.config.serverURL = process.env.SERVER_URL;
}

// get default values from the querystring if there's any
if (params.get("swma_calculation_type")) {
  window.config.swma_calculation_type = params.get(
    "swma_calculation_type"
  ) as SWMACalculationType;
}

if (params.get("swma_threshold_type")) {
  window.config.swma_threshold_type = params.get(
    "swma_threshold_type"
  ) as SWMAThresholdType;
}

if (params.get("swma_threshold")) {
  window.config.swma_threshold = parseInt(
    params.get("swma_threshold") || "5",
    10
  );
}

if (params.get("swma_window_size")) {
  window.config.swma_window_size = parseInt(
    params.get("swma_window_size") || "50",
    10
  );
}

if (params.get("swma_calculation_interval")) {
  window.config.swma_calculation_interval = parseInt(
    params.get("swma_calculation_interval") || "10",
    10
  );
}

const logHandler = (txt: string) => {
  const div = document.createElement("div");
  const pre = document.createElement("pre");
  pre.innerText = txt;
  div.appendChild(pre);
  logContentRef.appendChild(div);
};

// fill resolutions combobox
Object.keys(window.config.resolutions).forEach((key) => {
  resolutionsRef.options[resolutionsRef.options.length] = new Option(
    window.config.resolutions[key],
    key
  );
});

Object.keys(window.config.throttleData).forEach((key) => {
  throttleDDL.options[throttleDDL.options.length] = new Option(
    window.config.throttleData[parseInt(key)],
    key
  );
});

const plotConfig = {
  toImageButtonOptions: {
    format: "svg", // one of png, svg, jpeg, webp
    filename: "custom_image",
    width: 700,
    scale: 1, // Multiply title/legend/axis/canvas sizes by this factor
  },
  displayModeBar: true,
  scrollZoom: true,
  displaylogo: false,
  responsive: true,
} as Plotly.Config;

const plotLayout = {
  hovermode: "closest",
  margin: {
    r: 10,
    t: 40,
    b: 40,
    l: 50,
  },
  height: 400,
  width: 440,
  title: "",
  showlegend: true,
  legend: {
    x: 0,
    y: -0.3,
    orientation: "h",
  },
  grid: {
    rows: 1,
    columns: 1,
    pattern: "independent",
  },
  xaxis: {
    anchor: "y",
    type: "linear",
    showgrid: true,
    showticklabels: true,
    title: "Time (s)",
    rangemode: "tozero",
  },
  yaxis: {
    anchor: "x",
    showgrid: true,
    title: "Mbps",
    rangemode: "tozero",
  },
  font: {
    family: "sans-serif",
    size: 18,
    color: "#000",
  },
} as Plotly.Layout;
//plotlayout for latency
const plotLayoutLatency = {
  hovermode: "closest",
  margin: {
    r: 10,
    t: 40,
    b: 40,
    l: 50,
  },
  height: 400,
  width: 440,
  title: "",
  showlegend: true,
  legend: {
    x: 0,
    y: -0.3,
    orientation: "h",
  },
  grid: {
    rows: 1,
    columns: 1,
    pattern: "independent",
  },
  xaxis: {
    anchor: "y",
    type: "linear",
    showgrid: true,
    showticklabels: true,
    title: "Time (s)",
    rangemode: "tozero",
  },
  yaxis: {
    anchor: "x",
    showgrid: true,
    title: "Latency (ms)",
    rangemode: "tozero",
  },
  font: {
    family: "sans-serif",
    size: 18,
    color: "#000",
  },
} as Plotly.Layout;

const plotLayoutJitter = {
  hovermode: "closest",
  margin: {
    r: 10,
    t: 40,
    b: 40,
    l: 50,
  },
  height: 400,
  width: 440,
  title: "",
  showlegend: true,
  legend: {
    x: 0,
    y: -0.3,
    orientation: "h",
  },
  grid: {
    rows: 1,
    columns: 1,
    pattern: "independent",
  },
  xaxis: {
    anchor: "y",
    type: "linear",
    showgrid: true,
    showticklabels: true,
    title: "Time (s)",
    rangemode: "tozero",
  },
  yaxis: {
    anchor: "x",
    showgrid: true,
    title: "Jitter (ms)",
    rangemode: "tozero",
  },
  font: {
    family: "sans-serif",
    size: 18,
    color: "#000",
  },
} as Plotly.Layout;

//plotlayout for throughput
const plotLayoutThroughput = {
  hovermode: "closest",
  margin: {
    r: 10,
    t: 40,
    b: 40,
    l: 50,
  },
  height: 400,
  width: 440,
  title: "",
  showlegend: true,
  legend: {
    x: 0,
    y: -0.3,
    orientation: "h",
  },
  grid: {
    rows: 1,
    columns: 1,
    pattern: "independent",
  },
  xaxis: {
    anchor: "y",
    type: "linear",
    showgrid: true,
    showticklabels: true,
    title: "Time (s)",
    rangemode: "tozero",
  },
  yaxis: {
    anchor: "x",
    showgrid: true,
    title: "Throughput (Mbps)",
    rangemode: "tozero",
  },
  font: {
    family: "sans-serif",
    size: 18,
    color: "#000",
  },
} as Plotly.Layout;

const plotData = [
  {
    x: [] as number[],
    y: [] as number[],
    name: "Server ETP",
    mode: "markers",
    xaxis: "x",
    yaxis: "y",
    marker: {
      color: "black",
      size: 11,
      symbol: "cross-thin",
      line: {
        width: 3,
      },
    },
  },
  {
    x: [] as number[],
    y: [] as number[],
    name: "tc Rate",
    mode: "line",
    xaxis: "x",
    yaxis: "y",
    line: {
      color: "#0905ed",
      width: 3,
    },
  },
  {
    x: [] as number[],
    y: [] as number[],
    name: "SWMA",
    mode: "markers",
    xaxis: "x",
    yaxis: "y",
    marker: {
      color: "#b33dc6",
      size: 11,
      symbol: "x-thin",
      line: {
        width: 3,
        color: "red",
      },
    },
  },
  {
    x: [] as number[],
    y: [] as number[],
    name: "IFA",
    mode: "markers",
    xaxis: "x",
    yaxis: "y",
    marker: {
      color: "#037325",
      size: 11,
      symbol: "star-triangle-down",
    },
  },
  {
    x: [],
    y: [],
    name: "Active Bandwidth Test",
    mode: "markers",
    xaxis: "x",
    yaxis: "y",
    marker: {
      size: 7,
      color: "#27aeef",
    },
  },
] as any[];
const plotLatencyData = [
  {
    x: [] as number[],
    y: [] as number[],
    name: "Latency",
    mode: "line",
    xaxis: "x",
    yaxis: "y",
    line: {
      color: "#ff0000",
      width: 3,
    },
  },
] as any[];

const plotJitterData = [
  {
    x: [] as number[],
    y: [] as number[],
    name: "Latency",
    mode: "line",
    xaxis: "x",
    yaxis: "y",
    line: {
      color: "#838f35",
      width: 3,
    },
  },
] as any[];

const plotThroughputData = [
  {
    x: [] as number[],
    y: [] as number[],
    name: "Throughput",
    mode: "line",
    xaxis: "x",
    yaxis: "y",
    line: {
      color: "#0905ed",
      width: 3,
    },
  },
] as any[];

const plot = Plotly.newPlot(
  document.getElementById("plot") as HTMLDivElement,
  plotData,
  plotLayout,
  plotConfig
);
const plotLatency = Plotly.newPlot(
  document.getElementById("plot_latency") as HTMLDivElement,
  plotLatencyData,
  plotLayoutLatency,
  plotConfig
);
const plotJitter = Plotly.newPlot(
  document.getElementById("plot_jitter") as HTMLDivElement,
  plotJitterData,
  plotLayoutJitter,
  plotConfig
);
const plotThroughput = Plotly.newPlot(
  document.getElementById("plot_throughput") as HTMLDivElement,
  plotThroughputData,
  plotLayoutThroughput,
  plotConfig
);
const player = new Player({
  url: params.get("url") || window.config.serverURL,
  vid: vidRef,
  stats: statsRef,
  throttle: throttleRef,
  throttleDDLRef: throttleDDL,
  resolutions: resolutionsRef,
  activeBWTestRef: activeBWTestRef,
  continueStreamingRef: continueStreamingRef,
  categoryRef: categoryRef,
  activeBWAsset: window.config.activeBWAsset,
  activeBWTestInterval: window.config.activeBWTestInterval,
  autioStart: window.config.autoStart || true,
  logger: logHandler,
});

// expose player
window.player = player;

let timePassed = 0;
let playerRefreshInterval = 1000; // 1 second
const displayedHistory = 240; // 4 minutes
const plotStartDelay = 4000; // 4 seconds
const testDuration = window.config.testDuration || 0;

let plotTimer: NodeJS.Timer;

const startPlotting = () => {
  console.log("in startPlotting");
  plotTimer = setInterval(() => {
    if (!player.started || player.paused) {
      return;
    }
    timePassed += playerRefreshInterval;

    const currentSec = Math.round(timePassed / 1000);

    if (testDuration > 0 && currentSec === testDuration) {
      player.pauseOrResume(false);
      player.downloadStats().then((results) => {
        console.log("results", results);
      });
      return;
    }

    // save results by time
    // these will be downloaded after the test
    player.saveResultBySecond(
      "etp",
      player.serverBandwidth / 1000000 || 0,
      player.currCategory,
      player.throughputs.get("SWMALatency") || 0,
      player.throughputs.get("chunk"),
      currentSec
    );
    // player.saveResultBySecond('tcRate', player.tcRate || 0, currentSec);

    plotData.forEach((p) => (p.x as Plotly.Datum[]).push(currentSec));
    plotLatencyData.forEach((p) => (p.x as Plotly.Datum[]).push(currentSec));
    plotThroughputData.forEach((p) => (p.x as Plotly.Datum[]).push(currentSec));
    plotJitterData.forEach((p) => (p.x as Plotly.Datum[]).push(currentSec));
    (plotData[0].y as Plotly.Datum[]).push(player.serverBandwidth / 1000000);
    (plotData[1].y as Plotly.Datum[]).push(player.tcRate / 1000000);
    (plotLatencyData[0].y as Plotly.Datum[]).push(
      player.throughputs.get("avgSegmentLatency") || 0
    );
    (plotJitterData[0].y as Plotly.Datum[]).push(
      player.throughputs.get("avgSegmentJitter") || 0
    );
    (plotThroughputData[0].y as Plotly.Datum[]).push(
      (player.throughputs.get("chunk") || 0) / 1000
    );

    // show max 60 seconds
    if (plotData[0].x.length > displayedHistory) {
      plotData.forEach((item) => {
        (item.x as Plotly.Datum[]).splice(0, 1);
        (item.y as Plotly.Datum[]).splice(0, 1);
      });
    }

    const data_update = {
      x: Object.values(plotData).map((item) => item.x),
      y: Object.values(plotData).map((item) => item.y),
    } as Plotly.Data;

    const data_latency_update = {
      x: Object.values(plotLatencyData).map((item) => item.x),
      y: Object.values(plotLatencyData).map((item) => item.y),
    } as Plotly.Data;
    const data_jitter_update = {
      x: Object.values(plotJitterData).map((item) => item.x),
      y: Object.values(plotJitterData).map((item) => item.y),
    } as Plotly.Data;

    const data_throughput_update = {
      x: Object.values(plotThroughputData).map((item) => item.x),
      y: Object.values(plotThroughputData).map((item) => item.y),
    } as Plotly.Data;

    Plotly.update(
      document.getElementById("plot") as Plotly.Root,
      data_update,
      plotLayout
    );
    Plotly.update(
      document.getElementById("plot_latency") as Plotly.Root,
      data_latency_update,
      plotLayoutLatency
    );
    Plotly.update(
      document.getElementById("plot_jitter") as Plotly.Root,
      data_jitter_update,
      plotLayoutJitter
    );
    Plotly.update(
      document.getElementById("plot_throughput") as Plotly.Root,
      data_throughput_update,
      plotLayoutThroughput
    );
  }, playerRefreshInterval);
};

startRef.addEventListener("click", async (e) => {
  e.preventDefault();
  if (!player.started) {
    await player.start();
    if (player.started) {
      document
        .querySelectorAll("#controls :disabled")
        .forEach((e) => e.removeAttribute("disabled"));
      startRef.innerText = "Stop";
      setTimeout(() => startPlotting(), plotStartDelay);
    } else {
      alert("Error occurred in starting!");
    }
  } else {
    player.stop();
  }
});

liveRef.addEventListener("click", (e) => {
  e.preventDefault();
  player.goLive();
});

toggleLogRef.addEventListener("click", (e) => {
  const logEl = document.getElementById("log");
  if (!logEl) {
    return;
  }

  if (toggleLogRef.innerText === "Show Logs") {
    toggleLogRef.innerText = "Hide Logs";
    logEl.style.display = "block";
  } else {
    toggleLogRef.innerText = "Show Logs";
    logEl.style.display = "none";
  }
});

function playFunc(e: Event) {
  // Only fire once to restore pause/play functionality
  vidRef.removeEventListener("play", playFunc);
}

vidRef.addEventListener("play", playFunc);
vidRef.volume = 0.5;

// Try to autoplay but ignore errors on mobile; they need to click
// vidRef.play().catch((e) => console.warn(e))

// datagramRef.addEventListener("click", async (e) => {
//     e.preventDefault();
//     const baseUrl = params.get("url") || window.config.serverURL
//     const url = baseUrl + "/test-datagram"
//     console.log("Creating WebTransport connection to " + url)
//     const quic = new WebTransport(url)
//     await quic.ready

//     receiveDatagram(quic)

//     // Send datagrams to the server.
//     const writer = quic.datagrams.writable.getWriter()
//     let utf8Encode = new TextEncoder();
//     const data1 = "a b c d e f g h i j k l m n o p q r s t u v w x y z";
//     const data2 = "1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18";
//     for (let i = 1; i <= 20; i++) {
//         if (i % 2 !== 0) {
//             writer.write(utf8Encode.encode(data1));
//             console.log(`Sending Datagram ${i}: ${data1}`)
//         } else {
//             writer.write(utf8Encode.encode(data2));
//             console.log(`Sending Datagram ${i}: ${data2}`)
//         }
//     }
// })

// datagramRef2.addEventListener("click", async (e) => {
//     e.preventDefault();
//     const baseUrl = params.get("url") || window.config.serverURL
//     console.log("Creating WebTransport connection to " + baseUrl)
//     const quic = new WebTransport(baseUrl)
//     await quic.ready

//     receiveDatagram(quic)
// })

// stremRef2.addEventListener("click", async (e) => {
//     e.preventDefault();
//     const baseUrl = params.get("url") || window.config.serverURL
//     console.log("Creating WebTransport connection to " + baseUrl)
//     const quic = new WebTransport(baseUrl)
//     await quic.ready

//     const uni = await quic.createUnidirectionalStream();
//     // uni.getWriter
//     const streams = quic.incomingUnidirectionalStreams.getReader();

//     let count = 1
//     while (true) {
//         const {value, done} = await streams.read();
//         if (done) {
//             break;
//         }
//         // value is a Uint8Array.
//         let utf8Decode = new TextDecoder;
//         const reader = value.getReader();
//         const a = await reader.read()
//         // console.log(a)
//         console.log(`Received Stream ${count}: ${utf8Decode.decode(a.value)}`);
//         count++
//     }
// })

async function receiveDatagram(quic: WebTransport) {
  // Read datagrams from the server.
  const reader = quic.datagrams.readable.getReader();
  let count = 1;
  while (true) {
    const { value, done } = await reader.read();
    if (done) {
      break;
    }
    // value is a Uint8Array.
    let utf8Decode = new TextDecoder();
    console.log(`Received Datagram ${count}:${utf8Decode.decode(value)}`);
    // console.log(`Received Datagram ${count}:${value.slice(0,2)}${utf8Decode.decode(value.slice(2))}`);
    count++;
  }
}

// stremRef.addEventListener("click", async (e) => {
//     e.preventDefault();
//     const baseUrl = params.get("url") || window.config.serverURL
//     const url = baseUrl + "/test-stream"
//     console.log("Creating WebTransport connection to " + url)
//     const quic = new WebTransport(url)
//     await quic.ready
//     const stream = await quic.createBidirectionalStream();

//     receiveStream(stream.readable.getReader())

//     // Send streams to the server.
//     const writer = stream.writable.getWriter();
//     let utf8Encode = new TextEncoder();
//     const data1 = "abc";
//     const data2 = "def";
//     for (let i = 1; i <= 20; i++) {
//         if (i % 2 !== 0) {
//             writer.write(utf8Encode.encode(data1));
//             console.log(`Sending Stream ${i}: ${data1}`)
//         } else {
//             writer.write(utf8Encode.encode(data2));
//             console.log(`Sending Stream ${i}: ${data2}`)
//         }
//     }
// })

async function receiveStream(reader: ReadableStreamDefaultReader) {
  // Read streams from the server.
  let count = 1;
  while (true) {
    const { value, done } = await reader.read();
    if (done) {
      break;
    }
    // value is a Uint8Array.
    let utf8Decode = new TextDecoder();
    console.log(`Received Stream ${count}: ${utf8Decode.decode(value)}`);
    count++;
  }
}

// TODO: Not used, somehow different from https://developer.chrome.com/docs/capabilities/web-apis/webtransport#webtransportreceivestream
async function readFrom(receiveStream: any) {
  const reader = receiveStream;
  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    // value is a Uint8Array.
    let utf8Decode = new TextDecoder();
    console.log(utf8Decode.decode(value));
  }
}
