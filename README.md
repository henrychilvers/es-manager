# ES Manager
Look for, and optionally remove, unnecessary Elastic Search indexes (or indices if you prefer).

Unnecessary indexes include:
* Empty indexes
* Future indexes (default, based on a pattern of *-YYYY-MM-dd. i.e. "kpi-2018-10-11")
* Old indexes (indexes older than the given date)

### Technology Stack
```
* Go
* Elastic Search
```
### Build
```
$ go build
```
### Usage
```
$ es-manager -v [6|2] -url=<http://192.168.2.10:9200> -ip=<ind1,ind2> [-lt=YYYY-MM-dd|n] [-d]
```
#### Arguments
```
    -v      Version of ES at target URL (currently only 6 or 2 are supported)
    
    -url    ES URL (e.g. -url=http://192.168.2.10:9200)

    -ip     Comma eparated list of index name prefixes (e.g. -ip=kpi-,data-)

    -lt     Look for indexes before this cut-off date (not inclusive), or n number of days prior to today.
    
    -d      If given, delete the unnecessary indexes, else just look for them 
```
#### Pre-requisites
```
* Go 1.8+ (https://golang.org/dl/)
~~*Elastic Search 1.4.4~~
* Elastic Search 6.*
```
#### Development Pre-requisites
```
* Go compiler
* GoLand (or IDE of choice hopefully capable of Go syntax awareness)
```
Contributors
-----
* Henry Chilvers