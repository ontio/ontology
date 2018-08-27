## initialize && RUN
### 1.initialize(Run only once)
* ontology start.sh add Executable authority 
  * chmod +x /opt/gopath/test/start.sh
* initialize httpd settings and restart httpd service
  * sh /opt/start.sh IP (IP:External network IP of VM)
  ![avatar](azure_image/initialize.png)
* start ontology testmode service
  * cd /opt/gopath/test && nohup ./start.sh &
  ![avatar](azure_image/ontology.png)

### 2.NetworkSecurityGroup
* VM->SETTINGS->Networking->Inbound port rules
  * 80
  * 8000
  * 3306（Optional）
  * 8080
  ![avatar](azure_image/securityGroup.png)

## Using(IP:External network IP of VM)
* smartx:http://IP
 ![avatar](azure_image/smartx.png)
* explorer:http://IP:8000
 ![avatar](azure_image/explorer.png)

## Software infomation
### 1.mysql
* port:3306
* user:root
* passwd:123456
* database:explorer;ontscide
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
  * smartx Front end (/var/www/html/ont-sc-ide) :80
  * smartx back end  (/var/www/html/sc-project-ser):8080
  * explorer  (/var/www/html):8000 

### 4.java service
* explorer service 
  * path: /root/explorer 
  * port:8085
  * description:Provide page logic API 
* sync servicde 
  * path: /root/ontsynhandler 
  * port:10010
  * description:Synchronization block chain information
