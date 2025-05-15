const weatherData = {
  50: { description: '阳光充沛' },
  51: { description: '间有阳光' },
  52: { description: '短暂阳光' },
  53: { description: '间有阳光几阵骤雨' },
  54: { description: '短暂阳光有骤雨' },
  60: { description: '多云' },
  61: { description: '密云' },
  62: { description: '微雨' },
  63: { description: '雨' },
  64: { description: '大雨' },
  65: { description: '雷暴' },
  70: { description: '天色良好(晚)' },
  71: { description: '天色良好(晚)' },
  72: { description: '天色良好(晚)' },
  73: { description: '天色良好(晚)' },
  74: { description: '天色良好(晚)' },
  75: { description: '天色良好(晚)' },
  76: { description: '大致多云(晚)' },
  77: { description: '天色大致良好(晚)' },
  80: { description: '大风' },
  81: { description: '乾燥' },
  82: { description: '潮湿' },
  83: { description: '雾' },
  84: { description: '薄雾' },
  85: { description: '烟霞' },
  90: { description: '热' },
  91: { description: '暖' },
  92: { description: '凉' },
  93: { description: '冷' }
}

export function getWeatherDetails(code: number) {
  if (typeof code !== 'number') {
    console.error('Invalid type for weather code:', code)
    return ''
  }
  const description = weatherData[code]?.description
  if (!description) {
    console.error('No description found for weather code:', code)
    return ''
  }
  return description
}
