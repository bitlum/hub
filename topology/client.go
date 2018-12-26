package topology

// Topology responsibilities:
//
// 0. Ask lightning client for querying routes and adding metrics to
// channels and nodes, so that decision about better path could be made
// wiser.
//
// 1. Store historical changes of lighting network topology:
// 	1.1 Overall channel change over time.
//
// 2. Derive additional metrics about nodes healthiness.
// 	2.1 Percentage of time node being online.
// 	2.2 Spam rate of node.
// 	2.3 Last time node being active.
//
// 3. Derive anomalies:
//	3.1 Drastic network capacity increase / decrease.
//	3.2 Drastic particular nodes capacity increase / decrease.
//	3.3 Drastic online percentage change for nodes known to us.
//
// 4. Derive additional general metrics about network:
// 	4.1 Average channel time.
// 	4.2 Average channel open / close fee.
// 	4.3 Average channel routing fees.
type Topology struct{}
