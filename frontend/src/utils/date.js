/**
 * 时间格式化工具
 * 将UTC时间转换为北京时间 (UTC+8)
 */

/**
 * 格式化时间为北京时间
 * @param {string|Date} timeStr - 时间字符串或Date对象
 * @param {string} format - 格式，默认 'YYYY-MM-DD HH:mm:ss'
 * @returns {string} 格式化后的时间字符串
 */
export const formatBeijingTime = (timeStr, format = 'YYYY-MM-DD HH:mm:ss') => {
  if (!timeStr) return ''
  
  const date = new Date(timeStr)
  if (isNaN(date.getTime())) return ''

  // 转换为北京时间 (UTC+8)
  const beijingTime = new Date(date.getTime() + 8 * 60 * 60 * 1000)
  
  const year = beijingTime.getUTCFullYear()
  const month = String(beijingTime.getUTCMonth() + 1).padStart(2, '0')
  const day = String(beijingTime.getUTCDate()).padStart(2, '0')
  const hours = String(beijingTime.getUTCHours()).padStart(2, '0')
  const minutes = String(beijingTime.getUTCMinutes()).padStart(2, '0')
  const seconds = String(beijingTime.getUTCSeconds()).padStart(2, '0')

  return format
    .replace('YYYY', year)
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds)
}

/**
 * 格式化日期（不包含时间）
 */
export const formatDate = (timeStr) => {
  return formatBeijingTime(timeStr, 'YYYY-MM-DD')
}

/**
 * 格式化时间（不包含日期）
 */
export const formatTime = (timeStr) => {
  return formatBeijingTime(timeStr, 'HH:mm:ss')
}
