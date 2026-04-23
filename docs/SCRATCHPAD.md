# SCRATCHPAD
This scratchpad is for taking notes on how to develop specific functionalities

## Importing data
Data can be imported via csv, this is done via vendors. A vendor has a specific mapping to a parser.
These vendors have parsers that live in their respective domains (i.e. portfolio, cashflow etc.). The parser only gets called in the following situation:
- The user imports a file
- The file gets saved on disk
- A background job checks the imports that need to be done
- The background job sends the file to their respective parsers in the domain
- The transactions get created in their respective domain

## Calculating ROI on positions
We can infer the positions based on the transactions. When new transactions get uploaded we need to 
check when the last snapshots were made per position and if we can create positions from the new 
transactions or map them to existing positions. Positions should be mapped to the listing. A listing 
is a mapping towards a provider. A provider is an external data source that provides the market 
data we need.

## Gathering dailies from external systems
We have providers that we can create, these are mappings to urls and external systems. When 
gathering data from there we want to do that as efficiently as possible. Therefore we have listings. 
We only gather dailies whent the dailies service gets called. We then spawn a background task only 
when we need to update the dailies data. The listing is responsible for maintaining it has been 
updated. Updating the listing is an atomic operation (only one at the time) we lock the row.

## Cashflow

## Client updates via websockets
Via websockets we are able to notify clients that data has changed. When data has changed the client updates themselves by fetching/ refreshing the data. The websocket updates start by the notify (or other name) package listening to update events on the eventbus. When an update is received (portfolioRebuilt, TransactionsImported or something similar, bulk tag finished) it will sent a notification to the websocket of that user, while the socket is active. The api will periodically send a ping to the client and the client will respond with a pong (this will happen for the web socket of the account_id), every account_id has it's own socket. Every 10 seconds (after no update) the ping message will be sent. When no pong message is received after 3 tries the socket connection is marked as stale and is disposed. When a client connects it sends a open connection request. The api checks if there is a socket open for that account id. When there is no socket open it opens a new one else it routes the data through that socket. When a socket hasn't received an update message for 5 minutes it can also be marked as stale. When the client detects clicks again the socket will try and reconnect.