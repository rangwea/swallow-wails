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
  let month = zeroFill(value.getMonth() + 1);
  let day = zeroFill(value.getDate());
  let hour = zeroFill(value.getHours());
  let minute = zeroFill(value.getMinutes());
  let second = zeroFill(value.getSeconds());

  return value.getFullYear() +
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
}

function zeroFill(i) {
  if (i >= 0 && i <= 9) {
    return "0" + i;
  } else {
    return i;
  }
}
