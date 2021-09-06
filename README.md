# go-scheduler
Multi threaded tool to connect multiple databases, Looks up for events which are due for start/stop as per timestamp and process. 
It Pushes status and data to redis store, which is used by  control panel(separate app) to monitor the operations and also re-issue the failed commands.
