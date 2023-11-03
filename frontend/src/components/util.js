/**
 * get current date time, format: yyyy-MM-dd HH:MM:SS
 */
export function getCurrentTime() {
  return formatDateTime(new Date());
}

/**
 * format date time: yyyy-MM-dd HH:MM:SS
 */
export function formatDateTime(value) {
  var month = zeroFill(value.getMonth() + 1);
  var day = zeroFill(value.getDate());
  var hour = zeroFill(value.getHours());
  var minute = zeroFill(value.getMinutes());
  var second = zeroFill(value.getSeconds());

  var curTime =
    value.getFullYear() +
    "-" +
    month +
    "-" +
    day +
    " " +
    hour +
    ":" +
    minute +
    ":" +
    second;

  return curTime;
}

function zeroFill(i) {
  if (i >= 0 && i <= 9) {
    return "0" + i;
  } else {
    return i;
  }
}
