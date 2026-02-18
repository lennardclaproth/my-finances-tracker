/*
# PORTFOLIO

The portfolio package is responsible for calculating the performance of a portfolio.

It does a few things:

	- It maintains snapshots of the portfolio and accounts at different points in time.
	- It calculates the performance of the portfolio and accounts based on those snapshots.
	- It calculates the performance of the positions in the portfolio based on the snapshots.

A snapshot is essentially a freeze-frame of reality at a given moment:

	- How much money has been invested so far
	- What the portfolio is currently worth on the market
	- Whether the investor is up or down relative to their cost basis
	- How the account evolves over time

To make performance calculations easier, each snapshot maintains a set of derived fields:

## Snapshot Fields

### 1. Total Bought

The total amount of money deposited or invested into the account up to that snapshot.

This is usually a cumulative number and only changes when new capital is added.

Example:

	TotalBought = 1934.52€

### 2. Total Bought Net

The total invested amount minus transaction fees or costs.

This represents the true cost basis of the portfolio.

Formula:

	NetBought = TotalBought - Fees

Example:

	NetBought = 1934.52 - 13.36 = 1921.16€

### 3. Account Worth

The current market value of the portfolio at the snapshot timestamp.

This changes daily as asset prices move.

Example:

	AccountWorth = 1869.42€

### 4. Account Weight

A relative performance index starting at 100% on the first snapshot.

This answers:

	“How much has my account grown compared to day zero?”

Formula:

	AccountWeight = (AccountWorth / InitialAccountWorth) * 100

Example:

	InitialWorth = 1836.70
	AccountWeight = (1869.42 / 1836.70) * 100 = 101.78%

Meaning:

	The account is up +1.78% compared to the initial snapshot.

### 5. Delta (EUR)

The absolute gain or loss compared to the invested net amount.

Formula:

	DeltaEUR = AccountWorth - NetBought

Example:

	DeltaEUR = 1869.42 - 1921.16 = -51.74€

Meaning:

	The investor is still down 51.74€ relative to cost basis.

### 6. Delta Percentage

The same delta expressed as a percentage of the net invested amount.

Formula:

	DeltaPerc = (DeltaEUR / NetBought) * 100

Example:

	DeltaPerc = (-51.74 / 1921.16) * 100 = -2.69%

Meaning:

	The account is 2.69% below break-even.

### 7. Daily Delta (EUR)

The change in account worth compared to the previous snapshot.

This captures daily market movement.

Formula:

	DailyDeltaEUR = WorthToday - WorthYesterday

Example:

	DailyDeltaEUR = 1869.42 - 1836.70 = +32.72€

Meaning:

	The portfolio gained 32.72€ during that day.

### 8. Daily Delta Percentage

The daily change expressed as a percentage.

Formula:

	DailyDeltaPerc = (DailyDeltaEUR / WorthYesterday) * 100

Example:

	DailyDeltaPerc = (32.72 / 1836.70) * 100 = +1.78%

### 9. Account Weighted Performance

A performance metric adjusted for cash flows.

This answers:

	“How well did the investments perform, ignoring deposits or withdrawals?”

In finance this is closely related to time-weighted return.

If no cash flows occur between snapshots, weighted performance will closely match daily percentage change.

## What These Metrics Tell You

These snapshot fields allow the system to distinguish between:

	- Market performance (account worth changes)
	- Investor profitability (delta vs net invested)
	- Time evolution (daily movement)
	- True return independent of deposits (weighted performance)

In short:

	AccountWorth tells you where you are.
	Delta tells you whether you are winning or still climbing out of fees.
	DailyDelta tells you what the market did today.
	WeightedPerformance tells you what your strategy did, free of cash flow distortion.

The universe is noisy, markets are chaotic, and these fields are how we keep score
without fooling ourselves.
*/
package portfolio
