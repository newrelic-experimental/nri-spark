[![Experimental Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Experimental.png)](https://opensource.newrelic.com/oss-category/#experimental)

![GitHub forks](https://img.shields.io/github/forks/newrelic-experimental/nri-spark?style=social)
![GitHub stars](https://img.shields.io/github/stars/newrelic-experimental/nri-spark?style=social)
![GitHub watchers](https://img.shields.io/github/watchers/newrelic-experimental/nri-spark?style=social)

![GitHub all releases](https://img.shields.io/github/downloads/newrelic-experimental/nri-spark/total)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/newrelic-experimental/nri-spark)
![GitHub last commit](https://img.shields.io/github/last-commit/newrelic-experimental/nri-spark)
![GitHub Release Date](https://img.shields.io/github/release-date/newrelic-experimental/nri-spark)


![GitHub issues](https://img.shields.io/github/issues/newrelic-experimental/nri-spark)
![GitHub issues closed](https://img.shields.io/github/issues-closed/newrelic-experimental/nri-spark)
![GitHub pull requests](https://img.shields.io/github/issues-pr/newrelic-experimental/nri-spark)
![GitHub pull requests closed](https://img.shields.io/github/issues-pr-closed/newrelic-experimental/nri-spark)


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


## Installation

The integration can be deployed independently on linux 64 system or as a databricks integration using a notebook. The sections below suggests each.


###### Standalone deployment 

1. Create "*nr-spark-metric-settings.yml*" file in the the folder "*/etc/nri-spark-metric/*"  using the following format
    ```
    sparkmasterurl: "http://localhost:8080"  <== FQDN ofspark master URL
    clustername:  mylocalcluster             <== Name of the cluster
    insightsapikey: xxxx                     <== Insights api key
    pollinterval: 5                          <== Polling interval
    clustermode:                             <== Set mode to *spark_driver_mode* for Single Node clusters
    tags:                                    <== Additional tags to be added to metrics
      nr_sample_tag_org: newrelic_labs
      nr_sample_tag_practice: odp
    ```
2. Run the following command.
    > service nr-spark-metric start

3. Check for metrics in "Metric" event type in Insights




###### Databricks Init script creator notebook

>**This notebook and configuration is for reference purpose only, deployment should customize this to fulfill the needs**


1. Create a new notebook to deploy the cluster intialization script 
2. Copy the relevant script below. You do not need to set or touch the $DB_ values in the script, Databricks populates these for us. 
   a Optional : Based on cluster install mode, uncommment  SingleNodeCluster install , comment Standalone
   b Optional : Install infra agent, update with latest version
4. Replace **<Add your insights key>>** with your New Relic Insights Insert Key. 
5. Add/Remove/Update tags require in the tag section, sample tags are configured using *nr_sample_tag\** 
6. Run this notebook to create to deploy the new_relic_install.sh script in dbfs in configured folder.
7. Ensure the script is attached to your cluster and is listed in the notebooks of the cluster
8. Running this script will create the file at dbfs:/nr/nri-spark-metric.sh
9. Configure target cluster with the ***newrelic_install.sh*** cluster-scoped init script using the UI, Databricks CLI, or by invoking the Clusters API. This setting is found in Cluster configuration tab -> Advanced Options -> Init Scripts
10. Add dbfs:/nr/nri-spark-metric.sh and click add. 
11. Restart your cluster
12. Metrics should start reporting under the Metrics section in New Relic with the prefix of spark.X.X - you should get Job, Stage Executors and Stream metrics.

```
dbutils.fs.put("dbfs:/nr/nri-spark-metric.sh",""" 
#!/bin/sh
echo "Check if this is driver? $DB_IS_DRIVER"
echo "Spark Driver ip: $DB_DRIVER_IP"

#Create Cluster init script
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
  
  echo "Check if this is driver? $DB_IS_DRIVER"
  echo "Spark Driver ip: $DB_DRIVER_IP"
    
  # Optional install infra agent
    
  # echo "license_key: <<NR LICENSE KEY>>" | \$sudo tee -a /etc/newrelic-infra.yml         
  ##Download upstart based installer to /tmp. # update the linke with new relases
  #\$sudo wget https://download.newrelic.com/infrastructure_agent/linux/apt/pool/main/n/newrelic-infra/newrelic-infra_upstart_1.3.27_upstart_amd64.deb -P /tmp
  #\$sudo wget https://download.newrelic.com/infrastructure_agent/linux/apt/pool/main/n/newrelic-infra/newrelic-infra_upstart_1.9.0_upstart_amd64.deb -P /tmp
  ## Install NR Agent
  #\$sudo dpkg -i /tmp/newrelic-infra_upstart_1.3.27_upstart_amd64.deb
    

#Download nr-spark-metric integration
  \$sudo wget https://github.com/newrelic-experimental/nri-spark/releases/download/1.2.0/nri-spark-metric.tar.gz  -P /tmp


#Extract the contents to right place
  \$sudo tar -xvzf /tmp/nri-spark-metric.tar.gz -C /
  
  # Check which mode is the cluster running in  
  # Start of  SingleNodeCluster install , using "spark_driver_mode"', uncomment this section and comment out Standalone cluster
  #  echo '  > SingleNodeCluster, using "spark_driver_mode"'
  #  DB_DRIVER_PORT=\$(grep -i "CONF_UI_PORT" /tmp/driver-env.sh | cut -d'=' -f2)
  #  SPARK_CLUSTER_MODE='spark_driver_mode'
  # end of SingleNodeCluster install
    
  # Start of Standalone Cluster, use the below section 
    echo '  > Standalone cluster, using "spark_standalone_mode", waiting for master-params...'
    while [ -z \$is_available ]; do
      if [ -e "/tmp/master-params" ]; then
        DB_DRIVER_PORT=\$(cat /tmp/master-params | cut -d' ' -f2)
        SPARK_CLUSTER_MODE=''
        is_available=TRUE
      fi
      sleep 2
    done
  # end of Standalone Cluster section
  
#Configure nr-spark-metric-settings.yml file 

  echo "sparkmasterurl: http://\$DB_DRIVER_IP:\$DB_DRIVER_PORT
clustername: \$DB_CLUSTER_NAME
insightsapikey: NRII-XXXXXXXXXXXXXXXX	
pollinterval: 5
clustermode: \$SPARK_CLUSTER_MODE
tags:
  nr_sample_tag_org: newrelic_labs
  nr_sample_tag_practice: odp" > /etc/nri-spark-metric/nr-spark-metric-settings.yml

  echo ' > Configured  nr-spark-metric-settings.yml \n $(</etc/nri-spark-metric/nr-spark-metric-settings.yml)'

#Enable the service
 \$sudo systemctl enable nr-spark-metric.service

  #Start the service 
  
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

New Relic has open-sourced this project. This project is provided AS-IS WITHOUT WARRANTY OR DEDICATED SUPPORT. Issues and contributions should be reported to the project here on GitHub.
>We encourage you to bring your experiences and questions to the [Explorers Hub](https://discuss.newrelic.com) where our community members collaborate on solutions and new ideas.


## Contributing
We encourage your contributions to improve [project name]! Keep in mind when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.
If you have any questions, or to execute our corporate CLA, required if your contribution is on behalf of a company,  please drop us an email at opensource@newrelic.com.

## License
New Relic Infrastructure Integration for Apache Spark is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.

