var curDate = new Date(); 
var preDate = new Date(curDate.getTime() - 24*60*60*1000);  
var nextDate = new Date(curDate.getTime() + 24*60*60*1000);
console.log(curDate)
console.log(preDate)
console.log(nextDate)
