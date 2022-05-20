<h1>Mining-Opportunity-Cost-Calculator</h1>

Run this go program with the flags and you will get command line output stats as well as a generated `points.png` file generated in the root of this repo.

Note only the <code>cli</code> folder is fit to use for cli usage, the <code>pkg</code> folder is being used as a server to serve up the data/chart since eventually this may run somewhere and support a front end. If you want to run the server, head to the bottom of this page for instructions.

Not that your slush token does expire after some time, and using an expired token will results in the calls to fail.
Also note that using a slush token is optional, if you know your total amount of bitcoin mined just use the <code>-bitcoinMined</code> flag instead.

Flags

<ul>
<li><code>-token</code> slush pool token (https://slushpool.com/settings/access/) you can use this flag or the <code>-bitcoinMined</code> flag if you dont use slush</li>
<li><code>-startDate</code> mm/dd/yyyy format for start date of mining operation</li>
<li><code>-kwhPrice</code> price paid per kilowatt-hour</li>
<li><code>-watts</code> watts used by the miers</li>
<li><code>-electricCosts</code> if you know your total amount spent on electric, can use it here instead of kwhPrice and watts and uptimePercent</li>
<li><code>-uptimePercent</code> percent of time mining operation is online (expressed as an integer)</li>
<li><code>-fixedCosts</code> total costs of miners, hardware, and other operational fixed costs</li>
<li><code>-bitcoinMined</code> amount of bitcoin mined (whole bitcoin units not sats)</li>
<li><code>-messariApiKey</code> api key from messari.io for historical price data</li>
<li><code>-hideBitcoinOnGraph</code> Will hide bitcoin on y-axis of graph, good for opsec when sharing the image. <code>true</code> to hide, <code>false</code> to keep the figure displayed</li>
</ul>


Run this from within the `cli` folder.

Example with token: `go run main.go -token abc123 -startDate 01/01/2022 -kwhPrice .14 -watts 3300 -uptimePercent 98 -fixedCosts 7500 -hideBitcoinOnGraph=true`

Example without token and with known total electric costs: `go run main.go -bitcoinMined 0.420 -startDate 01/01/2022 -electricCosts 3400 -fixedCosts 7500`


Output (these are made up figures for the example here):
```
Bicoin current price: $35947.77
Days since start: 420.93
Average coins per day: 0.00040153
Dollarinos earned: $4931.38
Total electric costs: $2231.03
Percent paid off: 68.88%
Bitcoin percentage increase needed to be breakeven: 49.53%
Breakeven price: $63751.70
Expected more days until breakeven: 128.57
Total mining days (past + future) to breakeven: 549.50
Expected breakeven date: 11/21/2022


------------------------------------------------

Electric costs per day: $8.64
bitcoin mined: 0.155
AmericanHodl: 0.1500202135019427
Daily-DCA: 0.19892472898958916
Anti-Miner: 0.16783636213105427


------------------------------------------------

Percentage comparison of strategies versus mining. 

Daily-DCA: 21.56%
Anti-Miner: 2.65%
AmericanHodl: -9.08%

```

![Example output plot](example-points.png)

<h3>Lines Explained</h3>
<li><b>AmericanHodl</b> - This strategy is if on the first day you slam bought all the bitcoin with all the fiat. This fiat amount is the sum of your mining operations fixed costs plus all the costs in electricity usage</li>
<li><b>DCA</b> Short for "Dollar cost averaging" this strategy refers to taking the sum of the fixed and varialbe costs (electric), dividing this number by total number of days since mining started, and stacked that amount of dollars worth of bitcoin each day. (daily DCA strategy)</li>
<li><b>Anti-Miner</b> This strategy refers to the person who spend an equal number of dollars on bitcoin purchasing each time the miner spends money on a cost. So on the first day, they buy the amount of bitcoin (in fiat terms) equal to the amount for the mining operation's fixed cost setup. Each day after they purchase the amount of bitcoin equal to the amount the miner spent of electricity that day</li>
<li><b>Mined</b> This line represents total bitcoin mined. This tool does not yet support entering amount miner per day to generate a proper historical line, and really should be represented as a singular point all the way on the last day of the x-axis. However, that becomes visually hard to see and for optics I simply had it plot as the entire width of the axis.</li>


<h2>Server Instructions</h2>

This can be run as a server from the root of this repo. 

Bring up ther server with `go run main.go` 

There are two endpoint:  `/data` and `/chart` where `/data` will return the text data you would see if you used the CLI, and `/data` yields the chart the CLI also generates.

Ping `localhost:8080/data` or ``localhost:8080/data`` with a json body that may look something like:

```
{
  "bitcoinMined": 0.12345,
  "startDate": "06/30/2021",
  "fixedCosts": 5000,
  "hideBitcoinOnGraph": false,
  "electricCosts": 3000.45
}
```

Here's a curl command for example: 

```
curl -i -H "Accept: application/json" -H "Content-Type: application/json" -d '{"bitcoinMined": 0.12345, "startDate":"06/30/2021", "fixedCosts":5000, "hideBitcoinOnGraph":false, "kwhPrice": 0.1133, "watts": 3400, "uptimePercent": 99 }' -X POST http://localhost:8080/chart

```