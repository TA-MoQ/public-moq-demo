const plotConfig = {
    toImageButtonOptions: {
      format: 'svg', // one of png, svg, jpeg, webp
      filename: 'custom_image',
      width: 700,
      scale: 1, // Multiply title/legend/axis/canvas sizes by this factor
    },
    displayModeBar: true,
    scrollZoom: true,
    displaylogo: false,
    responsive: true,
  } as Plotly.Config;
  
  const plotLayout = {
    hovermode: 'closest',
    margin: {
      r: 10,
      t: 40,
      b: 40,
      l: 50,
    },
    height: 400,
    width: 440,
    title: '',
    showlegend: true,
    legend: {
      x: 0,
      y: -0.3,
      orientation: 'h',
    },
    grid: {
      rows: 1,
      columns: 1,
      pattern: 'independent',
    },
    xaxis: {
      anchor: 'y',
      type: 'linear',
      showgrid: true,
      showticklabels: true,
      title: 'Time (s)',
      rangemode: 'tozero',
    },
    yaxis: {
      anchor: 'x',
      showgrid: true,
      title: 'Mbps',
      rangemode: 'tozero',
    },
    font: {
      family: 'sans-serif',
      size: 18,
      color: '#000',
    },
  } as Plotly.Layout;
  //plotlayout for latency
  const plotLayoutLatency = {
    hovermode: 'closest',
    margin: {
      r: 10,
      t: 40,
      b: 40,
      l: 50,
    },
    height: 400,
    width: 440,
    title: '',
    showlegend: true,
    legend: {
      x: 0,
      y: -0.3,
      orientation: 'h',
    },
    grid: {
      rows: 1,
      columns: 1,
      pattern: 'independent',
    },
    xaxis: {
      anchor: 'y',
      type: 'linear',
      showgrid: true,
      showticklabels: true,
      title: 'Time (s)',
      rangemode: 'tozero',
    },
    yaxis: {
      anchor: 'x',
      showgrid: true,
      title: 'Latency (ms)',
      rangemode: 'tozero',
    },
    font: {
      family: 'sans-serif',
      size: 18,
      color: '#000',
    },
  } as Plotly.Layout;
  //plotlayout for throughput
  const plotLayoutThroughput = {
    hovermode: 'closest',
    margin: {
      r: 10,
      t: 40,
      b: 40,
      l: 50,
    },
    height: 400,
    width: 440,
    title: '',
    showlegend: true,
    legend: {
      x: 0,
      y: -0.3,
      orientation: 'h',
    },
    grid: {
      rows: 1,
      columns: 1,
      pattern: 'independent',
    },
    xaxis: {
      anchor: 'y',
      type: 'linear',
      showgrid: true,
      showticklabels: true,
      title: 'Time (s)',
      rangemode: 'tozero',
    },
    yaxis: {
      anchor: 'x',
      showgrid: true,
      title: 'Throughput (Mbps)',
      rangemode: 'tozero',
    },
    font: {
      family: 'sans-serif',
      size: 18,
      color: '#000',
    },
  } as Plotly.Layout;
  
  const plotData = [
    {
      x: [] as number[],
      y: [] as number[],
      name: 'Server ETP',
      mode: 'markers',
      xaxis: 'x',
      yaxis: 'y',
      marker: {
        color: 'black',
        size: 11,
        symbol: 'cross-thin',
        line: {
          width: 3,
        },
      },
    },
    {
      x: [] as number[],
      y: [] as number[],
      name: 'tc Rate',
      mode: 'line',
      xaxis: 'x',
      yaxis: 'y',
      line: {
        color: '#0905ed',
        width: 3,
      },
    },
    {
      x: [] as number[],
      y: [] as number[],
      name: 'SWMA',
      mode: 'markers',
      xaxis: 'x',
      yaxis: 'y',
      marker: {
        color: '#b33dc6',
        size: 11,
        symbol: 'x-thin',
        line: {
          width: 3,
          color: 'red',
        },
      },
    },
    {
      x: [] as number[],
      y: [] as number[],
      name: 'IFA',
      mode: 'markers',
      xaxis: 'x',
      yaxis: 'y',
      marker: {
        color: '#037325',
        size: 11,
        symbol: 'star-triangle-down',
      },
    },
    {
      x: [],
      y: [],
      name: 'Active Bandwidth Test',
      mode: 'markers',
      xaxis: 'x',
      yaxis: 'y',
      marker: {
        size: 7,
        color: '#27aeef',
      },
    },
  ] as any[];
  const plotLatencyData = [
    {
      x: [] as number[],
      y: [] as number[],
      name: 'Latency',
      mode: 'line',
      xaxis: 'x',
      yaxis: 'y',
      line: {
        color: '#ff0000',
        width: 3,
      },
    },
  ] as any[];
  
  const plotThroughputData = [
    {
      x: [] as number[],
      y: [] as number[],
      name: 'Throughput',
      mode: 'line',
      xaxis: 'x',
      yaxis: 'y',
      line: {
        color: '#0905ed',
        width: 3,
      },
    },
  ] as any[];

export {plotConfig, plotLayout, plotLayoutLatency, plotLayoutThroughput, plotData, plotLatencyData, plotThroughputData}