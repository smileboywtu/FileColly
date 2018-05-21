// package define file collector command line and
// inter communicate parameter
package colly

// AppConfigOption define command line args
type AppConfigOption struct {
	RedisHost string `yaml:"redis_host" flagName:"redishost" flagSName:"rh" flagDescribe:"Destination Cache Redis host" default:"127.0.0.1"`
	RedisPort int    `yaml:"redis_port" flagName:"redisport" flagSName:"rp" flagDescribe:"Destination Cache Redis port" default:"6379"`
	RedisDB   int    `yaml:"redis_db" flagName:"redisdb" flagSName:"rdb" flagDescribe:"Destination Cache Redis db" default:"0"`
	RedisPW   string `yaml:"redis_passwd" flagName:"redispw" flagSName:"rpwd" flagDescribe:"Destination Cache Redis password" default:""`

	DestinationRedisQueueName  string `yaml:"dest_queue" flagName:"dqname" flagSName:"dq" flagDescribe:"Destination Redis Queue name" default:"paas:fileserver:files"`
	DestinationRedisQueueLimit int    `yaml:"dest_queue_limit" flagName:"dqlimit" flagSName:"dql" flagDescribe:"Destination Redis Queue size limit" default:"3000"`

	// wait time in second before reading the file
	// this make sure the file is ready
	ReadWaitTime int `yaml:"read_wait_time" flagName:"rwtime" flagSName:"rwt" flagDescribe:"Wait time before file can be read" default:"2"`

	SenderMaxWorkers int `yaml:"max_reader" flagName:"readers" flagSName:"rworker" flagDescribe:"Max worker for reading file" default:"500"`
	ReaderMaxWorkers int `yaml:"max_sender" flagName:"senders" flagSName:"sworker" flagDescribe:"Max worker for sending file" default:"500"`

	// max size in bytes that a file be filtered
	FileMaxSize string `yaml:"file_limit" flagName:"limit" flagSName:"flimit" flagDescribe:"File size limit in human size" default:"200M"`

	// do not delete file after sent
	ReserveFile bool `yaml:"reserve_file" flagName:"reserve" flagSName:"keep" flagDescribe:"Keep file after sent" default:"false""`

	// cache time in second before delete file
	FileCacheTimeout int `yaml:"cache_timeout" flagName:"ctime" flagSName:"ct" flagDescribe:"File Cache timeout" default:"3600"`

	LogFileName string `yaml:"log_file" flagName:"lfile" flagSName:"log" flagDescribe:"File to write log" default:"sender.log"`

	// file watch directory
	CollectDirectory string `yaml:"collect_directory" flagName:"cdir" flagSName:"d" flagDescribe:"File collect directory" default:"/opt/files"`
}
