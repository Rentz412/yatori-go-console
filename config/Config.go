package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

type JSONDataForConfig struct {
	Setting Setting `json:"setting"`
	Users   []Users `json:"users"`
}
type EmailInform struct {
	Sw       int    `json:"sw"`
	SMTPHost string `json:"smtpHost" yaml:"SMTPHost"`
	SMTPPort string `json:"smtpPort" yaml:"SMTPPort"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type BasicSetting struct {
	CompletionTone int    `default:"1" json:"completionTone,omitempty" yaml:"completionTone"` //是否开启刷完提示音，0为关闭，1为开启，默认为1
	ColorLog       int    `json:"colorLog,omitempty" yaml:"colorLog"`                         //是否为彩色日志，0为关闭彩色日志，1为开启，默认为1
	LogOutFileSw   int    `json:"logOutFileSw,omitempty" yaml:"logOutFileSw"`                 //是否输出日志文件0代表不输出，1代表输出，默认为1
	LogLevel       string `json:"logLevel,omitempty" yaml:"logLevel"`                         //日志等级，默认INFO，DEBUG为找BUG调式用的，日志内容较详细，默认为INFO
	LogModel       int    `json:"logModel" yaml:"logModel"`                                   //日志模式，0代表以视频提交学时基准打印日志，1代表以一个课程为基准打印信息，默认为0
}
type AiSetting struct {
	AiType ctype.AiType `json:"aiType" yaml:"aiType"`
	AiUrl  string       `json:"aiUrl" yaml:"aiUrl"`
	Model  string       `json:"model"`
	APIKEY string       `json:"API_KEY" yaml:"API_KEY" mapstructure:"API_KEY"`
}

type ApiQueSetting struct {
	Url string `json:"url"`
}

type Setting struct {
	BasicSetting  BasicSetting  `json:"basicSetting" yaml:"basicSetting"`
	EmailInform   EmailInform   `json:"emailInform" yaml:"emailInform"`
	AiSetting     AiSetting     `json:"aiSetting" yaml:"aiSetting"`
	ApiQueSetting ApiQueSetting `json:"apiQueSetting" yaml:"apiQueSetting"`
}
type CoursesSettings struct {
	Name         string   `json:"name"`
	IncludeExams []string `json:"includeExams" yaml:"includeExams"`
	ExcludeExams []string `json:"excludeExams" yaml:"excludeExams"`
}

// 新增课程项结构体，支持课程ID和名称
type CourseItem struct {
	Name string `json:"name" yaml:"name"`
	ID   string `json:"id" yaml:"id"`
}

type CoursesCustom struct {
	VideoModel      int               `json:"videoModel" yaml:"videoModel"`         //观看视频模式
	AutoExam        int               `json:"autoExam" yaml:"autoExam"`             //是否自动考试
	ExamAutoSubmit  int               `json:"examAutoSubmit" yaml:"examAutoSubmit"` //是否自动提交试卷
	ExcludeCourses  []interface{}     `json:"excludeCourses" yaml:"excludeCourses"` // 改为interface{}以兼容新旧格式
	IncludeCourses  []interface{}     `json:"includeCourses" yaml:"includeCourses"` // 改为interface{}以兼容新旧格式
	CoursesSettings []CoursesSettings `json:"coursesSettings" yaml:"coursesSettings"`
}
type Users struct {
	AccountType   string        `json:"accountType" yaml:"accountType"`
	URL           string        `json:"url"`
	Account       string        `json:"account"`
	Password      string        `json:"password"`
	IsProxy       int           `json:"isProxy" yaml:"isProxy"` //是否代理IP
	CoursesCustom CoursesCustom `json:"coursesCustom" yaml:"coursesCustom"`
}

// 读取json配置文件
func ReadJsonConfig(filePath string) JSONDataForConfig {
	var configJson JSONDataForConfig
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &configJson)
	if err != nil {
		log.Fatal(err)
	}
	return configJson
}

// 自动识别读取配置文件
func ReadConfig(filePath string) JSONDataForConfig {
	var configJson JSONDataForConfig
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		log2.Print(log2.INFO, log2.BoldRed, "找不到配置文件或配置文件内容书写错误")
		log.Fatal(err)
	}
	err = viper.Unmarshal(&configJson)
	//viper.SetTypeByDefaultValue(true)
	viper.SetDefault("setting.basicSetting.logModel", 5)

	if err != nil {
		log2.Print(log2.INFO, log2.BoldRed, "配置文件读取失败，请检查配置文件填写是否正确")
		log.Fatal(err)
	}
	
	// 转换课程配置格式
	for i := range configJson.Users {
		user := &configJson.Users[i]
		user.CoursesCustom.IncludeCourses = convertCourseFormat(user.CoursesCustom.IncludeCourses)
		user.CoursesCustom.ExcludeCourses = convertCourseFormat(user.CoursesCustom.ExcludeCourses)
	}
	
	return configJson
}

// 转换课程配置格式，将字符串数组转换为CourseItem数组
func convertCourseFormat(courses []interface{}) []interface{} {
    var result []interface{}
    for _, course := range courses {
        switch v := course.(type) {
        case string:
            // 旧格式：字符串，转换为CourseItem
            result = append(result, CourseItem{Name: v, ID: ""})
        case map[interface{}]interface{}:
            // 新格式：对象，转换为CourseItem
            name := ""
            id := ""
            if n, ok := v["name"].(string); ok {
                name = n
            }
            if i, ok := v["id"].(string); ok {
                id = i
            }
            result = append(result, CourseItem{Name: name, ID: id})
        case CourseItem:
            // 如果已经是CourseItem，直接使用
            result = append(result, v)
        default:
            // 尝试处理其他可能的格式
            if str, ok := course.(string); ok {
                result = append(result, CourseItem{Name: str, ID: ""})
            } else {
                // 无法识别的格式，记录日志或忽略
                fmt.Printf("无法识别的课程格式: %v\n", course)
            }
        }
    }
    return result
}

// CmpCourse 比较是否存在对应课程,匹配上了则true，没有匹配上则是false
// 修改为支持课程ID和名称的匹配
func CmpCourse(courseName, courseId string, courseList []interface{}) bool {
    lg.Print(lg.DEBUG, "开始匹配课程: 名称=", courseName, ", ID=", courseId)
    
    for i, item := range courseList {
        switch v := item.(type) {
        case string:
            lg.Print(lg.DEBUG, "  配置项[", i, "](字符串): ", v)
            if v == courseName {
                lg.Print(lg.DEBUG, "  匹配成功 (字符串)")
                return true
            }
        case config.CourseItem:
            lg.Print(lg.DEBUG, "  配置项[", i, "](CourseItem): 名称=", v.Name, ", ID=", v.ID)
            // 如果配置中指定了课程ID，则必须完全匹配ID
            if v.ID != "" {
                if v.ID == courseId {
                    lg.Print(lg.DEBUG, "  匹配成功 (ID匹配)")
                    return true
                }
            } else {
                // 如果配置中没有指定ID，则按名称匹配
                if v.Name == courseName {
                    lg.Print(lg.DEBUG, "  匹配成功 (名称匹配)")
                    return true
                }
            }
        }
    }
    lg.Print(lg.DEBUG, "  没有匹配项")
    return false
}

func GetUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func StrToInt(s string) int {
	res, err := strconv.Atoi(s)
	if err != nil {
		return 0 // 其他错误处理逻辑
	}
	return res
}
