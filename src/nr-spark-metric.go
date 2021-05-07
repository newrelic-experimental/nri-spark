package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/newrelic/newrelic-telemetry-sdk-go/cumulative"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/wait"
)

var nr NewRelic

// NewRelic nr structure
type NewRelic struct {
	harvestor *telemetry.Harvester
	dc        *cumulative.DeltaCalculator
}

// ConfigStruct to hold settings required
type ConfigStruct struct {
	SparkMasterURL     string `yaml:"sparkmasterurl"`
	ClusterName        string `yaml:"clustername"`
	InsightsAPIKey     string `yaml:"insightsapikey"`
	MetricsURLOverride string `yaml:"metricsurloverride"`
	PollInterval       int    `yaml:"pollinterval"`
}

// ActiveApps  struct which contains list of active apps
type ActiveApps struct {
	ActiveApps []App `json:"activeapps"`
}

// App structure
type App struct {
	ID             string `json:"id"`
	Starttime      int64  `json:"starttime"`
	Name           string `json:"name"`
	Cores          int    `json:"cores"`
	User           string `json:"user"`
	Memoryperslave int    `json:"memoryperslave"`
	Submitdate     string `json:"submitdate"`
	State          string `json:"state"`
	Duration       int    `json:"duration"`
}

//SparkJob structure
type SparkJob struct {
	JobID               int    `json:"jobId"`
	Name                string `json:"name"`
	SubmissionTime      string `json:"submissionTime"`
	StageIds            []int  `json:"stageIds"`
	Status              string `json:"status"`
	NumTasks            int    `json:"numTasks"`
	NumActiveTasks      int    `json:"numActiveTasks"`
	NumCompletedTasks   int    `json:"numCompletedTasks"`
	NumSkippedTasks     int    `json:"numSkippedTasks"`
	NumFailedTasks      int    `json:"numFailedTasks"`
	NumKilledTasks      int    `json:"numKilledTasks"`
	NumCompletedIndices int    `json:"numCompletedIndices"`
	NumActiveStages     int    `json:"numActiveStages"`
	NumCompletedStages  int    `json:"numCompletedStages"`
	NumSkippedStages    int    `json:"numSkippedStages"`
	NumFailedStages     int    `json:"numFailedStages"`
}

//SparkStage structure
type SparkStage struct {
	Status                string `json:"status"`
	StageID               int    `json:"stageId"`
	AttemptID             int    `json:"attemptId"`
	NumTasks              int    `json:"numTasks"`
	NumActiveTasks        int    `json:"numActiveTasks"`
	NumCompleteTasks      int    `json:"numCompleteTasks"`
	NumFailedTasks        int    `json:"numFailedTasks"`
	NumKilledTasks        int    `json:"numKilledTasks"`
	NumCompletedIndices   int    `json:"numCompletedIndices"`
	ExecutorRunTime       int    `json:"executorRunTime"`
	ExecutorCPUTime       int64  `json:"executorCpuTime"`
	SubmissionTime        string `json:"submissionTime"`
	FirstTaskLaunchedTime string `json:"firstTaskLaunchedTime"`
	InputBytes            int    `json:"inputBytes"`
	InputRecords          int    `json:"inputRecords"`
	OutputBytes           int    `json:"outputBytes"`
	OutputRecords         int    `json:"outputRecords"`
	ShuffleReadBytes      int    `json:"shuffleReadBytes"`
	ShuffleReadRecords    int    `json:"shuffleReadRecords"`
	ShuffleWriteBytes     int    `json:"shuffleWriteBytes"`
	ShuffleWriteRecords   int    `json:"shuffleWriteRecords"`
	MemoryBytesSpilled    int    `json:"memoryBytesSpilled"`
	DiskBytesSpilled      int    `json:"diskBytesSpilled"`
	Name                  string `json:"name"`
	SchedulingPool        string `json:"schedulingPool"`
	RddIds                []int  `json:"rddIds"`
	////AccumulatorUpdates []interface{} `json:"accumulatorUpdates"`
	//KilledTasksSummary struct {
	//} `json:"killedTasksSummary"`
}

//SparkExecutor struct
type SparkExecutor struct {
	ID                string `json:"id"`
	HostPort          string `json:"hostPort"`
	IsActive          bool   `json:"isActive"`
	RddBlocks         int    `json:"rddBlocks"`
	MemoryUsed        int    `json:"memoryUsed"`
	DiskUsed          int    `json:"diskUsed"`
	TotalCores        int    `json:"totalCores"`
	MaxTasks          int    `json:"maxTasks"`
	ActiveTasks       int    `json:"activeTasks"`
	FailedTasks       int    `json:"failedTasks"`
	CompletedTasks    int    `json:"completedTasks"`
	TotalTasks        int    `json:"totalTasks"`
	TotalDuration     int    `json:"totalDuration"`
	TotalGCTime       int    `json:"totalGCTime"`
	TotalInputBytes   int    `json:"totalInputBytes"`
	TotalShuffleRead  int    `json:"totalShuffleRead"`
	TotalShuffleWrite int    `json:"totalShuffleWrite"`
	IsBlacklisted     bool   `json:"isBlacklisted"`
	MaxMemory         int64  `json:"maxMemory"`
	AddTime           string `json:"addTime"`
	MemoryMetrics     struct {
		UsedOnHeapStorageMemory   int64 `json:"usedOnHeapStorageMemory"`
		UsedOffHeapStorageMemory  int64 `json:"usedOffHeapStorageMemory"`
		TotalOnHeapStorageMemory  int64 `json:"totalOnHeapStorageMemory"`
		TotalOffHeapStorageMemory int64 `json:"totalOffHeapStorageMemory"`
	} `json:"memoryMetrics"`
	//BlacklistedInStages []interface{} `json:"blacklistedInStages"`
	//ExecutorLogs        struct {
	//	Stdout string `json:"stdout"`
	//	Stderr string `json:"stderr"`
	//} `json:"executorLogs,omitempty"`
}

// SparkStreamStats struct
type SparkStreamStats struct {
	BatchDuration               int     `json:"batchDuration"`
	NumReceivers                int     `json:"numReceivers"`
	NumActiveReceivers          int     `json:"numActiveReceivers"`
	NumInactiveReceivers        int     `json:"numInactiveReceivers"`
	NumTotalCompletedBatches    int     `json:"numTotalCompletedBatches"`
	NumRetainedCompletedBatches int     `json:"numRetainedCompletedBatches"`
	NumActiveBatches            int     `json:"numActiveBatches"`
	NumProcessedRecords         int     `json:"numProcessedRecords"`
	NumReceivedRecords          int     `json:"numReceivedRecords"`
	AvgInputRate                float64 `json:"avgInputRate"`
	AvgSchedulingDelay          int     `json:"avgSchedulingDelay"`
	AvgProcessingTime           int     `json:"avgProcessingTime"`
	AvgTotalDelay               int     `json:"avgTotalDelay"`
}

// Global variables
var configData ConfigStruct
var nameSpace = "apacheSpark"

func makeRequest(requestURL string) []byte {

	log.Debug("makeRequest : request :", requestURL)
	response, err := http.Get(requestURL)
	if err != nil {
		log.Error("request failed with error ", err)
		//log.Fatal(err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		log.Trace("makeRequest : response body:" + string(data))
		return data
	}
	return nil
}

func getStandaloneAppURL(masterUIURL string, appID string) string {

	// Once we have teh app page we need to find the following string to identify appUI Task Executor metrics page
	// <a href="http://localhost:4040">Application Detail UI</a>
	// Create the app URL using APPID and
	appUIURL := masterUIURL + "/app/" + "?" + "appId=" + appID
	log.Debug("getStandaloneAppURL : fetching getStandaloneAppURL")

	byteValue := makeRequest(appUIURL)
	var appDetailUI = ""

	if byteValue != nil {
		// WE cannot find the App detail URL from api, use a hack
		appDetailUI = GetInnerSubstring(string(byteValue), "a href=\"", "\">Application Detail UI")
		log.Debug("getStandaloneAppURL : Application Detail UI" + appDetailUI)
	} else {
		log.Debug("getStandaloneAppURL : unable to obtain Application Detail URL")
	}
	return appDetailUI
}

// GetInnerSubstring gets us data
func GetInnerSubstring(str string, prefix string, suffix string) string {
	var beginIndex, endIndex int
	beginIndex = strings.Index(str, prefix)
	if beginIndex == -1 {
		beginIndex = 0
		endIndex = 0
	} else if len(prefix) == 0 {
		beginIndex = 0
		endIndex = strings.Index(str, suffix)
		if endIndex == -1 || len(suffix) == 0 {
			endIndex = len(str)
		}
	} else {
		beginIndex += len(prefix)
		endIndex = strings.Index(str[beginIndex:], suffix)
		if endIndex == -1 {
			if strings.Index(str, suffix) < beginIndex {
				endIndex = beginIndex
			} else {
				endIndex = len(str)
			}
		} else {
			if len(suffix) == 0 {
				endIndex = len(str)
			} else {
				endIndex += beginIndex
			}
		}
	}

	return str[beginIndex:endIndex]
}

//
func initStandalone(masterUIURL string, activeAppMap map[string]App) {

	var activeApps ActiveApps
	masterUIJSONURL := masterUIURL + "/json/"
	log.Debug("initStandalone : querying masterUI")
	byteValue := makeRequest(masterUIJSONURL)

	if byteValue != nil {
		if err := json.Unmarshal(byteValue, &activeApps); err != nil {
			log.Error("initStandalone : unable to get process active apps in json data")
			//log.Fatal(err)
		} else {
			// populate map that contains active app ID and App detail URL
			log.Debug("initStandalone : activeApps ", len(activeApps.ActiveApps))

			for i := 0; i < len(activeApps.ActiveApps); i++ {
				log.Debug("initStandalone : app ID: " + activeApps.ActiveApps[i].ID)
				log.Debug("initStandalone : app name: " + activeApps.ActiveApps[i].Name)
				appDetailURL := getStandaloneAppURL(masterUIURL, activeApps.ActiveApps[i].ID)
				activeAppMap[appDetailURL] = activeApps.ActiveApps[i]
			}
		}
	} else {
		log.Info("initStandalone : unable to get data from " + masterUIJSONURL)
	}

}

func getActiveApps(activeAppMap map[string]App) {

	// Get the running apps adn thier Task metrics URL from the master UI
	var masterUIurl = configData.SparkMasterURL

	// Do first for Standalone.
	initStandalone(masterUIurl, activeAppMap)
	//TDB  mesos/YARN

}

// Populate Job Metrics
func populateJobMetrics(appID string, appTaskURL string, tags map[string]interface{}) {
	var sparkJobs []SparkJob

	//http://localhost:4040/api/v1/applications/app-20190605145937-0017/jobs
	appTaskJobURL := appTaskURL + "/api/v1/applications/" + appID + "/jobs/"

	byteValue := makeRequest(appTaskJobURL)

	if err := json.Unmarshal(byteValue, &sparkJobs); err != nil {
		log.Error("populateJobMetrics : error processing json: ", err)
		return
		//fmt.Errorf("fatal error processing json: %s ", err)
	}

	// Get reflected value

	for i := 0; i < len(sparkJobs); i++ {
		e := reflect.ValueOf(&sparkJobs[i]).Elem()

		jobtags := make(map[string]interface{})
		setTags("spark.job.", e, tags, jobtags)
		log.Debug("populateJobMetrics : setting metrics for job:", sparkJobs[i].JobID, ", attributes: ", jobtags)

		setMetrics("spark.job.", e, jobtags)
	}
}

// Populate Stage Metrics
func populateStageMetrics(appID string, appTaskURL string, tags map[string]interface{}) {
	var sparkStages []SparkStage

	//http://localhost:4040/api/v1/applications/app-20190605145937-0017/stages
	appTaskStageURL := appTaskURL + "/api/v1/applications/" + appID + "/stages"

	byteValue := makeRequest(appTaskStageURL)

	if err := json.Unmarshal(byteValue, &sparkStages); err != nil {
		log.Error("populateStageMetrics : error processing json:  ", err)
		return
		//fmt.Errorf("fatal error processing json: %s  ", err)
	}
	for i := 0; i < len(sparkStages); i++ {
		// Get reflected value

		e := reflect.ValueOf(&sparkStages[i]).Elem()
		stagetags := make(map[string]interface{})

		setTags("spark.stage.", e, tags, stagetags)
		log.Debug("populateStageMetrics : setting metrics for stage:", sparkStages[i].StageID, ", attributes: ", stagetags)

		setMetrics("spark.stage.", e, stagetags)
	}

}

func populateExecutorMetrics(appID string, appTaskURL string, tags map[string]interface{}) {
	var sparkExecutors []SparkExecutor

	//http://localhost:4040/api/v1/applications/app-20190605145937-0017/executors
	appTaskStageURL := appTaskURL + "/api/v1/applications/" + appID + "/executors"

	byteValue := makeRequest(appTaskStageURL)
	//println(string(byteValue))
	if err := json.Unmarshal(byteValue, &sparkExecutors); err != nil {
		log.Error("populateExecutorMetrics : error processing json ", err)
		return
		//fmt.Errorf("fatal error processing json: %s  ", err)
	}

	for i := 0; i < len(sparkExecutors); i++ {
		// Get reflected value
		e := reflect.ValueOf(&sparkExecutors[i]).Elem()
		executortags := make(map[string]interface{})
		setTags("spark.executor.", e, tags, executortags)
		log.Debug("populateExecutorMetrics : setting metrics for executor:", sparkExecutors[i].ID, ", attributes: ", executortags)
		setMetrics("spark.executor.", e, executortags)
	}
}
func populateStreamingMetrics(appID string, appTaskURL string, tags map[string]interface{}) {

	var sparkStreamStats SparkStreamStats

	//http://localhost:4040/api/v1/applications/app-20190605145937-0017/executors
	appTaskStageURL := appTaskURL + "/api/v1/applications/" + appID + "/streaming/statistics"

	byteValue := makeRequest(appTaskStageURL)

	err := json.Unmarshal(byteValue, &sparkStreamStats)
	if err != nil {
		log.Warn("populateStreamingMetrics : error processing json: ", err)
		return
	}

	// Get reflected value
	e := reflect.ValueOf(&sparkStreamStats)
	streamingtags := make(map[string]interface{})

	setTags("spark.stream.", e, tags, streamingtags)
	setMetrics("spark.stream.", e, streamingtags)

}

func populateStorageMetrics(activeAppMap map[string]string) {}

func SplitToString(a []int, sep string) string {
	if len(a) == 0 {
		return ""
	}

	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(v)
	}
	return strings.Join(b, sep)
}

func setTags(prefix string, e reflect.Value, tags map[string]interface{}, metricTags map[string]interface{}) {

	for k, v := range tags {
		metricTags[k] = v
	}

	for i := 0; i < e.NumField(); i++ {

		var mname string
		mname = prefix + e.Type().Field(i).Name
		// populate string metrics as tags
		switch n := e.Field(i).Interface().(type) {
		case string:
			// Add this in tags
			log.Trace("setTags : adding tags ", mname, "=", n)
			metricTags[mname] = n
		case []int:
			metricTags[mname] = SplitToString(n, ",")
		default:
			//log.Debug("setTags :Skipping tags")
		}

	}
}

func setMetrics(prefix string, e reflect.Value, tags map[string]interface{}) {

	//set the time for one set
	mtime := time.Now()

	for i := 0; i < e.NumField(); i++ {

		var mvalue float64
		var mname string

		if e.Field(i).Kind() == reflect.Struct {
			log.Trace("encountered nested structure ")
			setMetrics(prefix, e.Field(i), tags)
			continue
		}

		mname = prefix + e.Type().Field(i).Name
		//varValue := e.Field(i).Interface()

		switch n := e.Field(i).Interface().(type) {
		case int:
			mvalue = float64(n)
		case int64:
			mvalue = float64(n)
		case uint64:
			mvalue = float64(n)
		case float64:
			mvalue = n
		case bool:
			mvalue = float64(0)
			if n {
				mvalue = float64(1)
			}
		default:
			log.Trace("setMetrics :skipping metric: ", n, mname, mvalue)
		}

		log.Debug("setMetrics : recording metric: ", mname, "=", mvalue)

		/*if counter, ok := nr.dc.CountMetric(mname, tags, mvalue, mtime); ok {
			nr.harvestor.RecordMetric(counter)
		}*/

		nr.harvestor.RecordMetric(telemetry.Gauge{
			Timestamp:  mtime,
			Value:      mvalue,
			Name:       strings.ToLower(mname),
			Attributes: tags})

	}
}
func initHarvestor() error {
	//Get the configuration data about the spark instance we're going to query.
	var err error

	err = readConfig()
	if err != nil {
		return err
	}

	nr.harvestor, err = telemetry.NewHarvester(telemetry.ConfigAPIKey(configData.InsightsAPIKey),
		telemetry.ConfigCommonAttributes(map[string]interface{}{
			"spark.clusterName": configData.ClusterName,
		}),
		func(cfg *telemetry.Config) {
			cfg.MetricsURLOverride = configData.MetricsURLOverride
		},
	)

	if err != nil {
		return fmt.Errorf("initHarvestor : unable to connect to newrelic %v", err)
	}
	nr.dc = cumulative.NewDeltaCalculator()
	return nil

}

//Spark specific helper functions defined below.
func readConfig() error {
	viper.SetConfigName("nr-spark-metric-settings")
	if value, exists := os.LookupEnv("NRSPARK_CONFIG"); exists {
		viper.AddConfigPath(value)
	} else {
		viper.AddConfigPath(".")
	}

	log.Info("readConfig : Reading Config ", viper.ConfigFileUsed())
	err := viper.ReadInConfig()
	if err != nil {
		// We will not be able to use it without config file access
		return fmt.Errorf("unable to read config: %v", err)
	}
	//viper.Debug()
	err = viper.Unmarshal(&configData)
	if err != nil {
		return fmt.Errorf("readConfig: Unable to unmarshall config: %v", err)
	}
	log.Info("readConfig : ClusterName :  " + configData.ClusterName)
	log.Info("readConfig : SparkMasterURL :  " + configData.SparkMasterURL)
	log.Info("readConfig : MetricsURLOverride :  " + configData.MetricsURLOverride)

	return nil
}

func populateSparkMetrics() {
	// Initilize maps
	activeApps := make(map[string]App)

	// Get all the active Applications by querying masterUI
	// We also need to populate the TaskMetric URL so we can query the actual Task metrics
	getActiveApps(activeApps)
	log.Info("populateSparkMetrics : active applications: ", len(activeApps))

	// run through all the active apps to fetch the App data
	for appTaskURL, app := range activeApps {
		//populate common tags for each app
		tags := make(map[string]interface{})
		tags["spark.app.Name"] = app.Name
		tags["spark.app.ID"] = app.ID
		tags["spark.app.Cores"] = app.Cores
		tags["spark.app.Starttime"] = app.Starttime
		tags["spark.app.State"] = app.State
		tags["spark.app.Submitdate"] = app.Submitdate
		tags["spark.app.User"] = app.User
		tags["spark.app.Memoryperslave"] = app.Memoryperslave
		log.Debug("populateSparkMetrics : adding universal tags: ", tags)

		populateJobMetrics(app.ID, appTaskURL, tags)
		populateStageMetrics(app.ID, appTaskURL, tags)
		populateExecutorMetrics(app.ID, appTaskURL, tags)
		populateStreamingMetrics(app.ID, appTaskURL, tags)
	}

}
func main() {

	log.Info("Starting nr-spark-metric")

	var loglevel string
	flag.StringVar(&loglevel, "loglevel", "info", "log level : info/debug/trace")

	var logfile string
	flag.StringVar(&logfile, "logfile", "", "log file location")

	flag.Parse()

	switch loglevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)

	log.Info("Log level : ", loglevel)

	if logfile != "" {
		log.Info("Log file : ", logfile)
		f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			// Cannot open log file. Logging to stderr
			log.Error("Unable to open log file", err)
		} else {
			log.SetOutput(f)
		}
	}

	err := initHarvestor()
	if err != nil {
		log.Fatal("nr-spark-metric.main: unable to initilize harvestor ", err.Error())
		os.Exit(1)
	}

	pollInterval := time.Duration(configData.PollInterval) * time.Second
	wait.PollInfinite(pollInterval, func() (bool, error) {
		populateSparkMetrics()
		log.Info("harvesting metrics ..")
		nr.harvestor.HarvestNow(context.Background())
		return false, nil
	})

}
