# product requirements doc for caddy/iyo

## functions

| id | Description | version | Remarks |
| --- | --- | --- |--- |
|  | caddy serves as reverse proxy for http(s) | 1.0 | |
|  | ACL: 1 iyo organization get access to caddy directive | 1.0 | |
|  | ACL: multiple iyo organizations get access to caddy directive | 1.1 | |
|  | track which user & from which organization accessed which page & when (logfile) | 1.0 | |
|  | https to http proxy works using letsencrypt certificate | 1.0 | |
|  | proxy while adding org & username to path (e.g. https://myserver/sshaccess/ goes to https://backendserver/sshaccess/$orgname/$username & other tricks for adding info to proxied url| 1.0 | |


## tested integrations

| id | Description | version | Remarks |
| --- | --- | --- |--- |
|  | ssh exposed over https behind caddy proxy | 1.0 | |
|  | jwt gets proxied to remote | 1.0 | |

## possible results 

| id | Description | version | Remarks |
| --- | --- | --- |--- |
|  | can track how long people where looking at content on website (fileserver) | 1.0 | |
|  | see which user looks at what content where | 1.0 | |
