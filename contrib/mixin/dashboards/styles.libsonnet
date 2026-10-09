// Shared visualization defaults from the reference dashboard. Arrays are replaced whole.
{
  timeseries: {
    defaults: {
      color: {
        mode: 'palette-classic',
      },
      custom: {
        axisBorderShow: false,
        axisCenteredZero: false,
        axisColorMode: 'text',
        axisLabel: '',
        axisPlacement: 'auto',
        barAlignment: 0,
        drawStyle: 'line',
        fillOpacity: 10,
        gradientMode: 'opacity',
        hideFrom: {
          legend: false,
          tooltip: false,
          viz: false,
        },
        insertNulls: false,
        lineInterpolation: 'linear',
        lineWidth: 1,
        pointSize: 4,
        scaleDistribution: {
          type: 'linear',
        },
        showPoints: 'never',
        spanNulls: false,
        stacking: {
          group: 'A',
          mode: 'none',
        },
        thresholdsStyle: {
          mode: 'off',
        },
      },
      thresholds: {
        mode: 'absolute',
        steps: [
          {
            color: 'green',
            value: null,
          },
        ],
      },
      unit: 'percentunit',
    },
    options: {
      legend: {
        calcs: [
          'mean',
          'max',
          'lastNotNull',
        ],
        displayMode: 'table',
        placement: 'bottom',
        showLegend: true,
      },
      tooltip: {
        hideZeros: false,
        mode: 'multi',
        sort: 'desc',
      },
    },
  },
  stat: {
    defaults: {
      color: {
        mode: 'thresholds',
      },
      thresholds: {
        mode: 'absolute',
        steps: [
          {
            color: 'text',
            value: null,
          },
        ],
      },
      unit: 'none',
    },
    options: {
      colorMode: 'value',
      graphMode: 'none',
      justifyMode: 'auto',
      orientation: 'auto',
      percentChangeColorMode: 'standard',
      reduceOptions: {
        calcs: [
          'lastNotNull',
        ],
        fields: '',
        values: false,
      },
      showPercentChange: false,
      textMode: 'auto',
      wideLayout: true,
    },
  },
  table: {
    defaults: {
      color: {
        mode: 'thresholds',
      },
      custom: {
        align: 'auto',
        cellOptions: {
          type: 'auto',
        },
        filterable: false,
        inspect: false,
        minWidth: 60,
      },
      thresholds: {
        mode: 'absolute',
        steps: [
          {
            color: 'green',
            value: null,
          },
        ],
      },
    },
    options: {
      cellHeight: 'sm',
      enablePagination: false,
      footer: {
        countRows: false,
        fields: '',
        reducer: [
          'sum',
        ],
        show: false,
      },
      showHeader: true,
      sortBy: [
        {
          desc: true,
          displayName: 'Used',
        },
      ],
    },
  },
  'state-timeline': {
    defaults: {
      color: {
        mode: 'thresholds',
      },
      custom: {
        fillOpacity: 90,
        hideFrom: {
          legend: false,
          tooltip: false,
          viz: false,
        },
        insertNulls: false,
        lineWidth: 0,
        spanNulls: false,
      },
      max: 1,
      min: 0,
      thresholds: {
        mode: 'absolute',
        steps: [
          {
            color: 'red',
            value: null,
          },
          {
            color: 'green',
            value: 1,
          },
        ],
      },
      unit: 'none',
    },
    options: {
      alignValue: 'left',
      legend: {
        displayMode: 'list',
        placement: 'bottom',
        showLegend: false,
      },
      mergeValues: false,
      rowHeight: 1,
      showValue: 'never',
      tooltip: {
        hideZeros: false,
        mode: 'single',
        sort: 'none',
      },
    },
  },
}
