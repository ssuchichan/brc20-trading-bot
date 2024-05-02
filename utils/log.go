package utils

import (
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// func init() {

// 	log.SetReportCaller(true)
// 	log.SetFormatter(&logrus.TextFormatter{
// 		FullTimestamp: true,
// 	})

// 	fileName := "logs/system.log"
// 	logDir := "logs"
// 	if _, err := os.Stat(logDir); os.IsNotExist(err) {
// 		err := os.Mkdir(logDir, 0755)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	logWriter, err := rotatelogs.New(
// 		fileName+".%Y%m%d",
// 		rotatelogs.WithLinkName(fileName),         // 生成软链，指向最新日志文件
// 		rotatelogs.WithMaxAge(7*24*time.Hour),     // 设置最大保存时间(7天)
// 		rotatelogs.WithRotationTime(24*time.Hour), // 设置日志切割时间间隔(1天)
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer logWriter.Close()

// 	log.SetOutput(logWriter)
// }

func GetLogger() *logrus.Logger {
	return log
}

var logScan = logrus.New()

func GetScanLogger() *logrus.Logger {
	logScan.SetReportCaller(true)
	logScan.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return log
}
