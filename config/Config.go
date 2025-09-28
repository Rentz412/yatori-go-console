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
    viper.SetDefault("setting.basicSetting.logModel", 5)

    if err != nil {
        log2.Print(log2.INFO, log2.BoldRed, "配置文件读取失败，请检查配置文件填写是否正确")
        log.Fatal(err)
    }
    
    log.Printf("原始解析的配置: %+v", configJson)
    
    // 转换课程配置格式
    for i := range configJson.Users {
        user := &configJson.Users[i]
        log.Printf("处理用户[%d]: %s", i, user.Account)
        log.Printf("  原始IncludeCourses: %v", user.CoursesCustom.IncludeCourses)
        log.Printf("  原始ExcludeCourses: %v", user.CoursesCustom.ExcludeCourses)
        
        user.CoursesCustom.IncludeCourses = convertCourseFormat(user.CoursesCustom.IncludeCourses)
        user.CoursesCustom.ExcludeCourses = convertCourseFormat(user.CoursesCustom.ExcludeCourses)
        
        log.Printf("  转换后IncludeCourses: %v", user.CoursesCustom.IncludeCourses)
        log.Printf("  转换后ExcludeCourses: %v", user.CoursesCustom.ExcludeCourses)
    }
    
    return configJson
}

// 转换课程配置格式，将字符串数组转换为CourseItem数组
// 转换课程配置格式，将字符串数组转换为CourseItem数组
func convertCourseFormat(courses []interface{}) []interface{} {
    var result []interface{}
    
    // 使用标准库的log输出，因为此时日志系统可能还未初始化
    log.Printf("开始转换课程配置，输入: %v (类型: %T)", courses, courses)
    
    for i, course := range courses {
        log.Printf("处理课程项[%d]: %v (类型: %T)", i, course, course)
        
        switch v := course.(type) {
        case string:
            // 旧格式：字符串，转换为CourseItem
            log.Printf("  识别为字符串格式: %s", v)
            result = append(result, CourseItem{Name: v, ID: ""})
        case map[interface{}]interface{}:
            // 新格式：对象，转换为CourseItem
            log.Printf("  识别为map[interface{}]interface{}格式: %v", v)
            name := ""
            id := ""
            if n, ok := v["name"].(string); ok {
                name = n
            }
            if i, ok := v["id"].(string); ok {
                id = i
            }
            log.Printf("  提取信息: 名称=%s, ID=%s", name, id)
            result = append(result, CourseItem{Name: name, ID: id})
        case map[string]interface{}:
            // 新格式：对象，转换为CourseItem
            log.Printf("  识别为map[string]interface{}格式: %v", v)
            name := ""
            id := ""
            if n, ok := v["name"]; ok {
                name = fmt.Sprintf("%v", n)
            }
            if i, ok := v["id"]; ok {
                id = fmt.Sprintf("%v", i)
            }
            log.Printf("  提取信息: 名称=%s, ID=%s", name, id)
            result = append(result, CourseItem{Name: name, ID: id})
        case CourseItem:
            // 如果已经是CourseItem，直接使用
            log.Printf("  已经是CourseItem格式: 名称=%s, ID=%s", v.Name, v.ID)
            result = append(result, v)
        default:
            // 尝试处理其他可能的格式
            log.Printf("  无法识别的格式，尝试转换")
            if str, ok := course.(string); ok {
                log.Printf("  成功转换为字符串: %s", str)
                result = append(result, CourseItem{Name: str, ID: ""})
            } else {
                // 最后尝试，使用fmt.Sprintf转换为字符串
                log.Printf("  使用fmt.Sprintf转换: %v", course)
                result = append(result, CourseItem{Name: fmt.Sprintf("%v", course), ID: ""})
            }
        }
    }
    
    log.Printf("转换课程配置完成，输出: %v", result)
    for i, item := range result {
        if courseItem, ok := item.(CourseItem); ok {
            log.Printf("  输出项[%d]: 名称=%s, ID=%s", i, courseItem.Name, courseItem.ID)
        } else {
            log.Printf("  输出项[%d]: %v (类型: %T)", i, item, item)
        }
    }
    
    return result
}

// CmpCourse 比较是否存在对应课程,匹配上了则true，没有匹配上则是false
// 修改为支持课程ID和名称的匹配
// CmpCourse 比较是否存在对应课程,匹配上了则true，没有匹配上则是false
func CmpCourse(courseName, courseId string, courseList []interface{}) bool {
    log.Printf("开始匹配课程: 名称=%s, ID=%s", courseName, courseId)
    log.Printf("配置列表: %v", courseList)
    
    for i, item := range courseList {
        switch v := item.(type) {
        case string:
            log.Printf("  配置项[%d](字符串): %s", i, v)
            if v == courseName {
                log.Printf("  匹配成功 (字符串)")
                return true
            }
        case CourseItem:
            log.Printf("  配置项[%d](CourseItem): 名称=%s, ID=%s", i, v.Name, v.ID)
            // 如果配置中指定了课程ID，则必须完全匹配ID
            if v.ID != "" {
                if v.ID == courseId {
                    log.Printf("  匹配成功 (ID匹配)")
                    return true
                } else {
                    log.Printf("  ID不匹配: 配置ID=%s, 课程ID=%s", v.ID, courseId)
                }
            } else {
                // 如果配置中没有指定ID，则按名称匹配
                if v.Name == courseName {
                    log.Printf("  匹配成功 (名称匹配)")
                    return true
                } else {
                    log.Printf("  名称不匹配: 配置名称=%s, 课程名称=%s", v.Name, courseName)
                }
            }
        default:
            log.Printf("  配置项[%d]未知类型: %v (类型: %T)", i, item, item)
        }
    }
    log.Printf("  没有匹配项")
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
