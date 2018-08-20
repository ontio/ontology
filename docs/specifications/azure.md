## Software introduction 
### 1.mysql
* port:3306
* user:root
* passwd:123456
* operation :systemctl start/stop/restart mysqld

### 2.ontology
* path:/opt/gopath/test
* operation:cd /opt/gopath/test && ./start.sh
* Log:/opt/gopath/test/Log
* wallet passwd:123456
* port:20334,20335,20336

### 3.httpd
* operation:systemctl start/stop/status/restart httpd
* port:80,8080,8000

### 4.java service
* /root/explorer port:8085
* /root/ontsynhandler port:10010


## initialize && RUN
### 1.initialize(Run only once)
```javascript
chmod +x /opt/gopath/test/ontology
sh /opt/start.sh IP (IP:External network IP of VM)
cd /opt/gopath/test 
nohup ./start.sh &
```
### 2.NetworkSecurityGroup
* VM->SETTINGS->Networking->Inbound port rules
  * 80
  * 8000
  * 3306（Optional）
  * 8080

## Using(IP:External network IP of VM)
* smartx:http://IP
* explorer:http://IP:8000
