[![Experimental Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Experimental.png)](https://opensource.newrelic.com/oss-category/#experimental)

<img alt="GitHub all releases" src="https://img.shields.io/github/downloads/newrelic-experimental/nri-spark/total?style=plastic">

# New Relic integration for Apache Spark 

This New Relic  standalone integration polls the Apache Spark [REST API](https://spark.apache.org/docs/latest/monitoring.html#rest-api) for metrics and pushes them into New Relic  using Metrics API
It uses the New Relic [Telemetry sdk for go](https://github.com/newrelic/newrelic-telemetry-sdk-go)

## Installation

*Requires Apache Spark runnning in standalone mode (YARN and mesos not yet supported)*

1. Download the latest package from Release.
2. Install the NR Spark Metric plugin plugin using the following command 
    > sudo tar -xvzf /tmp/nri-spark-metric.tar.gz -C /

    The following files will be installed 
    ```
    /etc/nri-spark-metric/
    /etc/nri-spark-metric/nr-spark-metric
    /etc/systemd/system/
    /etc/systemd/system/nr-spark-metric.service
    /etc/init/nr-spark-metric.conf
    ```


## Getting Started

The integration can be deployed independently on linux 64 system or as a databricks integration using a notebook. The sections below suggests each.


###### Standalone deployment 

1. Create "*nr-spark-metric-settings.yml*" file in the the folder "*/etc/nri-spark-metric/*"  using the following format
    ```
    sparkmasterurl: "http://localhost:8080"  <== FQDN ofspark master URL
    clustername:  mylocalcluster             <== Name of the cluster
    insightsapikey: xxxx                     <== Insights api key
    pollinterval: 5                          <== Polling interval
    workspace: myworkspace-name              <== Name Of Your Databricks Workspace (static value)
    environment: production                  <== Name Of Your Environment (static value)
    ```
2. Run the following command.
    > service nr-spark-metric start

3. Check for metrics in "Metric" event type in Insights




###### Databricks Init script creator notebook



1. Create a new notebook to deploy the cluster intialization script. 
2. Copy the script below in. You only need to make one change (step 3). You do not need to set or touch the $DB_ values in the script, Databricks populates these for us. 
3. Replace **<Add your insights key>>** with your New Relic Insights Insert Key. You can also adjust Workspace name (If using databricks for a static reference) and environment (production / staging / test) to help you dimensionalise multiple Databricks environments
4. Run this notebook to create to deploy the new_relic_install.sh script in dbfs in configured folder.
5. Ensure the script is attached to your cluster and is listed in the notebooks of the cluster
6. Running this script will create the file at dbfs:/nr/nri-spark-metric.sh
7. Configure target cluster with the ***newrelic_install.sh*** cluster-scoped init script using the UI, Databricks CLI, or by invoking the Clusters API. This setting is found in Cluster configuration tab -> Advanced Options -> Init Scripts
8. Add dbfs:/nr/nri-spark-metric.sh and click add. 
9. Restart your cluster
10. Metrics should start reporting under the Metrics section in New Relic with the prefix of spark.X.X - you should get Job, Stage Executors and Stream metrics.

```
dbutils.fs.put("dbfs:/nr/nri-spark-metric.sh",""" 
#!/bin/sh
echo "Check if this is driver? $DB_IS_DRIVER"
echo "Spark Driver ip: $DB_DRIVER_IP"

# Create Cluster init script
cat <<EOF >> /tmp/start_spark-metric.sh

#!/bin/sh

if [ \$DB_IS_DRIVER ]; then
  # Root user detection
  if [ \$(echo "$UID") = "0" ];                                      
  then                                                                     
    sudo=''                                                                
  else
    sudo='sudo'                                                        
  fi

  #Download nr-spark-metric integration
  \$sudo wget https://github.com/newrelic-experimental/nri-spark/releases/download/1.0.0/nri-spark-metric.tar.gz  -P /tmp


  #extract the contents to right place
  \$sudo tar -xvzf /tmp/nri-spark-metric.tar.gz -C /

  # GRAB PORT for masterUI
   while [ -z \$isavailable ]; do
    if [ -e "/tmp/master-params" ]; then
      DB_DRIVER_PORT=\$(cat /tmp/master-params | cut -d' ' -f2)
      isavailable=TRUE
    fi
    sleep 2
  done
  
   # configure spark instances
  echo "sparkmasterurl: http://\$DB_DRIVER_IP:\$DB_DRIVER_PORT
clustername: \$DB_CLUSTER_ID
insightsapikey: <<Add your insights key>>
pollinterval: 5 
workspace: myworkspace-name
environment: production" > /etc/nri-spark-metric/nr-spark-metric-settings.yml

   #enable 
 \$sudo systemctl enable nr-spark-metric.service
 
 #start the service 
 \$sudo systemctl start nr-spark-metric.service
 \$sudo start nr-spark-metric

fi
EOF

# Start 
if [ \$DB_IS_DRIVER ]; then
  chmod a+x /tmp/start_spark-metric.sh
  /tmp/start_spark-metric.sh >> /tmp/start_spark-metric.log 2>&1 & disown
fi

""",True)
```

## Support

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. Like all official New Relic open source projects, there's a related Community topic in the New Relic Explorers Hub. You can find this project's topic/threads here:


## Contributing
We encourage your contributions to improve [project name]! Keep in mind when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.
If you have any questions, or to execute our corporate CLA, required if your contribution is on behalf of a company,  please drop us an email at opensource@newrelic.com.

## License
New Relic Infrastructure Integration for Apache Spark is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.

