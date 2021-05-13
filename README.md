# Asterisk Outgoing Call API

Starts an outgoing call based on provided paramters according to the https://wiki.asterisk.org/wiki/display/AST/Asterisk+Call+Files call file format. View https://www.voip-info.org/asterisk-auto-dial-out/ for more details about the call file format. I created this to make my phone ring when my cell phone rings via a tasker profile. Should be useful for many additional things though.

# Accepted parameters

- token: The API token to authenticate with the server.
- channel: Channel to use for the call.
- caller_id: Caller ID, Please note: It may not work if you do not respect the format: CallerID: “Some Name” <1234>
- wait_time: Seconds to wait for an answer. Default is 45.
- max_retries: Number of retries before failing (not including the initial attempt, e.g. 0 = total of 1 attempt to make the call). Default is 0.
- retry_time: Seconds between retries, Don’t hammer an unavailable phone. The default is 300 (5 min).
- account: Set the account code to use.
- application: Asterisk Application to run (use instead of specifying context, extension and priority).
- data: The options to be passed to application.
- context: Context in extensions.conf
- extension: Extension definition in extensions.conf
- priority: Priority of extension to start with.
- set_var: Set of variables to set in url query format.
- archive: Yes/No – Move to subdir “outgoing_done” with “Status: value”, where value can be Completed, Expired or Failed.
- schedule: Schedule call for a later date/time. Can be natrual language input as parsed by https://github.com/olebedev/when 