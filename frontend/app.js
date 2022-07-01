const http = require('http');

const fs = require('fs')

const server = http.createServer((req, res) => {
  res.writeHead(200, { 'content-type': 'text/html' })
  fs.createReadStream('index.html').pipe(res)
})

server.listen(process.env.PORT || 3000)
// const data = JSON.stringify({
//   "bitcoinMined": 0.165, "startDate": "04/01/2021", "fixedCosts": 5638, "hideBitcoinOnGraph": false, "kwhPrice": 0.10628, "watts": 3205, "uptimePercent": 100
// });

// const options = {
//   hostname: 'ec2-18-215-175-235.compute-1.amazonaws.com',
//   port: 8080,
//   path: '/data',
//   method: 'POST',
//   headers: {
//     'Content-Type': 'application/json',
//     'Content-Length': data.length,
//   },
// };

// const req = http.request(options, res => {
//   console.log(`statusCode: ${res.statusCode}`);

//   res.on('data', d => {
//     process.stdout.write(d);
//   });
// });

// req.on('error', error => {
//   console.error(error);
// });

// req.write(data);
// req.end();