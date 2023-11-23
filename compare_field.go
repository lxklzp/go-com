package main

import (
	"fmt"
	"go-com/config"
	"go-com/global"
	"regexp"
	"strings"
)

func compareFieldJava() {
	str := `requestID; // 请求id
monitoringTime;        // 监控时间
notifyID;              // 实时告警trap定义的OID
alarmCSN;              // 1告警的网络流水号，唯一标识一条告警
category;              // 2告警种类，取值范围如下：1：故障2：恢复3：事件4：确认5：反确认9：变更
occurTime;             // 3告警发生时间
mOName;                // 4设备名称
productID;             // 5产品系列标识
nEType;                // 6设备类型标识
nEDevID;               // 7设备唯一标识
devCsn;                // 8告警的设备流水号，同一种网元内部唯一
alarmID;               // 9告警 ID，用来区分同一种类型设备的告警类型
alarmType;             // 10告警类型,取值范围见文档1：电源系统2：环境系统等..
alarmLevel;            // 11告警级别1：紧急2：重要3：次要4：警告6：恢复
restore;               // 12告警恢复标志1：已恢复2：未恢复
confirm;               // 13告警确认标志1：已确认2：未确认
extendInfo;            // 27告警扩展信息，主要包含了告警的定位信息。
probablecause;         // 28告警发生原因
objectInstanceType;    // 47对象实例类型。默认为 0
clearCategory;         // 48告警清除类别。默认为 0。
clearType;             // 49告警清除类型。1：正常清除2：复位清除3：手动清除4：配置清除5：相关性清除
serviceAffectFlag;     // 50影响服务的标记。0：否1：是
addionalInfo;          // 51附加信息
sameNumber;            // 31-45相同的数据
stateReference;        //告警头相关信息
alarmCounts; //告警数目
`
	regField, _ := regexp.Compile(`(.+)\n`)
	regPart, _ := regexp.Compile(`(\w+)\s*;\s*//\s*(.+)`)
	tmp := regField.FindAllStringSubmatch(str, -1)
	var fieldList [][]string
	var goCode string
	for _, line := range tmp {
		if line[0] == "" {
			continue
		}
		fieldList = append(fieldList, regPart.FindAllStringSubmatch(line[1], -1)[0])
	}

	str = `{
  "addionalInfo": "",
  "alarmCSN": "8439192",
  "alarmCounts": 0,
  "alarmID": "26007",
  "alarmLevel": "2",
  "alarmType": "15",
  "category": "1",
  "clearCategory": "1",
  "clearType": "0",
  "confirm": "2",
  "devCsn": "220578",
  "extendInfo": "机架号\u003d0, 位置号\u003d0, 框号\u003d0, 槽号\u003d4, 模块号\u003d11, 告警原因\u003d注册刷新频率超限",
  "mOName": "GAABAC01",
  "monitoringTime": "1.3.6.1.2.1.1.3.0\u003d246days,16:47:29.44",
  "nEDevID": "NE\u003d281",
  "nEType": "9178",
  "notifyID": "1.3.6.1.4.1.2011.2.15.2.4.3.3.0.1",
  "objectInstanceType": "0",
  "occurTime": "2023-09-14 18:24:25",
  "probablecause": "信令DoS攻击 ",
  "productID": "2",
  "requestID": "972788867",
  "restore": "2",
  "sameNumber": "16-",
  "serviceAffectFlag": "0",
  "stateReference": "StateReference[msgID\u003d0,pduHandle\u003dPduHandle[972788867],securityEngineID\u003dnull,securityModel\u003dnull,securityName\u003dpublic,securityLevel\u003d1,contextEngineID\u003dnull,contextName\u003dnull,retryMsgIDs\u003dnull], "
}`
	regFieldRaw, _ := regexp.Compile(`"(\w+)":`)
	fieldListRaw := regFieldRaw.FindAllStringSubmatch(str, -1)

	for _, line := range fieldListRaw {
		exist := false
		for _, l2 := range fieldList {
			if line[1] == l2[1] {
				exist = true
				break
			}
		}
		if !exist {
			global.Log.Fatal(line)
		}
	}

	global.Log.Info(goCode)
}

func compareField() {
	str := `"parentID" IS '父告警ID';

"groupID" IS '关联组号';

"relatedStatus" IS '告警关联状态';

"SRC_ALARMTITLE" IS '告警标题';

"SRC_PERCEIVEDSEVERITY" IS '告警级别';

"SRC_EQUIPMENTNAME" IS '设备名称';

"SRC_PROVINCE" IS '省代码';

"SRC_ALARMTYPE" IS '告警类型';

"SRC_IPADDRESS" IS 'IP地址';

"SRC_CLEARANCEREPORTFLAG" IS '清除状态';

"SRC_EQUIPMENTCODE" IS '设备代码';

"SRC_LOCATIONINFO" IS '定位信息';

"SRC_ALARMCODE" IS '告警码';

"SRC_NECLASS" IS '设备类型';

"SRC_ALARMSUBTYPE" IS '告警子类型';

"SRC_VENDOR" IS '设备厂商';

"SRC_CITY" IS '市';

"SRC_NETTYPE" IS '网络类型';

"SRC_ADDITIONALTEXT" IS '原始报文';

"SRC_INFO1" IS '附加信息1';

"SRC_INFO2" IS '附加信息2';

"SRC_INFO3" IS '附加信息3';

"SRC_INFO4" IS '附加信息4';

"SRC_INFO5" IS '附加信息5';

"SRC_INFO6" IS '附加信息6';

"SRC_INFO7" IS '附加信息7';

"SRC_INFO8" IS '附加信息8';

"SRC_INFO9" IS '附加信息9';

"IAS_ORDERSTATUS" IS '工单状态';

"IAS_ALARMCAUSE" IS '原因';

"IAS_IMPCATINFO" IS '影响';

"IAS_CLEARTIME" IS '清除时间';

"IAS_ALARMID" IS '告警ID';

"IAS_TALLY" IS '压缩次数';

"IAS_IS_MAINTENANCE" IS '养护';

"IAS_ORDERNO" IS '工单号';

"IAS_IMPACTLIST" IS '业务影响列表';

"IAS_LASTTIME" IS '最后一次发生时间';

"SRC_IS_TEST" IS '是否是测试 0 否 1 是';
`
	regField, _ := regexp.Compile(`(.+)\n`)
	regPart, _ := regexp.Compile(`"(\w+)"\s+IS\s+'(.+)'`)
	tmp := regField.FindAllStringSubmatch(str, -1)
	var fieldList [][]string
	var goCode string
	for _, line := range tmp {
		if line[0] == "" {
			continue
		}
		fieldList = append(fieldList, regPart.FindAllStringSubmatch(line[1], -1)[0])
	}

	str = `{
    "IAS_LASTTIME": "2023-11-22 20:04:43",
    "SRC_IS_TEST": "0",
    "IAS_IMPACTLIST": "",
    "SRC_ALARMTITLE": "check_sda_util_mgs",
    "relatedStatus": 0,
    "SRC_PERCEIVEDSEVERITY": 5,
    "SRC_INFO9": "",
    "SRC_EQUIPMENTNAME": "MGS-cascaded-nova-compute10.144.216.139",
    "SRC_PROVINCE": "521000000000000000000001",
    "SRC_ALARMTYPE": "",
    "parentID": "",
    "SRC_IPADDRESS": "10.144.216.139",
    "SRC_CLEARANCEREPORTFLAG": 1,
    "SRC_EQUIPMENTCODE": "10.144.216.139",
    "SRC_LOCATIONINFO": "azoneName=cn-gz1a,region=贵州,service=MGS-Management,component=MGS-cascaded-nova-compute,HOST_IP=10.144.216.139,HOST_NAME=GZ-AZ01-MGS-S3-PSZOM02-CNA-041",
    "IAS_SEDNTIME": "2023-11-22 20:16:01",
    "SRC_ALARMCODE": "CMC_9000001380",
    "IAS_MAINTENANCE_NO": "",
    "SRC_INFO8": "",
    "SRC_INFO7": "",
    "SRC_INFO6": "",
    "SRC_NECLASS": "MGS-Management",
    "SRC_INFO5": "",
    "SRC_INFO4": "10.144.216.139",
    "SRC_INFO3": "贵州",
    "SRC_INFO2": "cn-gz1",
    "IAS_ORDERNO": "",
    "SRC_INFO1": "HONOR_POOL",
    "SRC_ALARMSUBTYPE": "",
    "groupID": "",
    "SRC_VENDOR": "HW",
    "SRC_CITY": "贵州",
    "IAS_FIRSTTIME": "2023-11-22 20:04:43",
    "SRC_NETTYPE": "CMP-TYY",
    "IAS_ALARMCAUSE": "计算节点系统盘/dev/sda的磁盘IO使用率持续高，请检查是否存在异常读写！",
    "IAS_IMPCATINFO": "",
    "IAS_CLEARTIME": "2023-11-22 20:15:43",
    "IAS_ALARMID": "17168398",
    "IAS_TALLY": 1,
    "IAS_IS_MAINTENANCE": 0,
    "SRC_ADDITIONALTEXT": "CloudService=MGS-Management,Service=MGS-nova-compute,MicroService=MGS-cascaded-nova-compute,NativeMeDn=10.144.216.139,regionId=cn-gz1,ItemName=diskio.util,InstanceInfo=sda3,trigger=metric=Diskio@diskio.util:value.last(0)>98,currentValue=Diskio@diskio.util:last()>98,value:97.31,mk:sda3",
    "IAS_ORDERSTATUS": 0
}`
	regFieldRaw, _ := regexp.Compile(`"(\w+)":`)
	fieldListRaw := regFieldRaw.FindAllStringSubmatch(str, -1)

	for _, line := range fieldListRaw {
		exist := false
		for _, l2 := range fieldList {
			if line[1] == l2[1] {
				exist = true
				break
			}
		}
		if !exist {
			global.Log.Fatal(line)
		}
	}

	global.Log.Info(goCode)
}

func fieldToCodeJava() {
	str := `
requestID; // 请求id
monitoringTime;        // 监控时间
notifyID;              // 实时告警trap定义的OID
alarmCSN;              // 1告警的网络流水号，唯一标识一条告警
category;              // 2告警种类，取值范围如下：1：故障2：恢复3：事件4：确认5：反确认9：变更
occurTime;             // 3告警发生时间
mOName;                // 4设备名称
productID;             // 5产品系列标识
nEType;                // 6设备类型标识
nEDevID;               // 7设备唯一标识
devCsn;                // 8告警的设备流水号，同一种网元内部唯一
alarmID;               // 9告警 ID，用来区分同一种类型设备的告警类型
alarmType;             // 10告警类型,取值范围见文档1：电源系统2：环境系统等..
alarmLevel;            // 11告警级别1：紧急2：重要3：次要4：警告6：恢复
restore;               // 12告警恢复标志1：已恢复2：未恢复
confirm;               // 13告警确认标志1：已确认2：未确认
extendInfo;            // 27告警扩展信息，主要包含了告警的定位信息。
probablecause;         // 28告警发生原因
objectInstanceType;    // 47对象实例类型。默认为 0
clearCategory;         // 48告警清除类别。默认为 0。
clearType;             // 49告警清除类型。1：正常清除2：复位清除3：手动清除4：配置清除5：相关性清除
serviceAffectFlag;     // 50影响服务的标记。0：否1：是
addionalInfo;          // 51附加信息
sameNumber;            // 31-45相同的数据
stateReference;        //告警头相关信息
alarmCounts; //告警数目
`
	regField, _ := regexp.Compile(`(.+)\n`)
	regPart, _ := regexp.Compile(`(\w+)\s*;\s*//\s*(.+)`)
	tmp := regField.FindAllStringSubmatch(str, -1)
	var goCode string
	for _, line := range tmp {
		if line[0] == "" {
			continue
		}
		tmp2 := regPart.FindAllStringSubmatch(line[1], -1)[0]
		goCode += fmt.Sprintf("%s       string `gorm:\"column:%s;comment:%s\" json:\"%s\"`\n", strings.ToUpper(tmp2[1][:1])+tmp2[1][1:], tmp2[1], tmp2[2], tmp2[1])
	}
	global.Log.Info(goCode)
}

func fieldToCode() {
	str := `{
    "IAS_LASTTIME": "2023-11-22 20:04:43",
    "SRC_IS_TEST": "0",
    "IAS_IMPACTLIST": "",
    "SRC_ALARMTITLE": "check_sda_util_mgs",
    "relatedStatus": 0,
    "SRC_PERCEIVEDSEVERITY": 5,
    "SRC_INFO9": "",
    "SRC_EQUIPMENTNAME": "MGS-cascaded-nova-compute10.144.216.139",
    "SRC_PROVINCE": "521000000000000000000001",
    "SRC_ALARMTYPE": "",
    "parentID": "",
    "SRC_IPADDRESS": "10.144.216.139",
    "SRC_CLEARANCEREPORTFLAG": 1,
    "SRC_EQUIPMENTCODE": "10.144.216.139",
    "SRC_LOCATIONINFO": "azoneName=cn-gz1a,region=贵州,service=MGS-Management,component=MGS-cascaded-nova-compute,HOST_IP=10.144.216.139,HOST_NAME=GZ-AZ01-MGS-S3-PSZOM02-CNA-041",
    "IAS_SEDNTIME": "2023-11-22 20:16:01",
    "SRC_ALARMCODE": "CMC_9000001380",
    "IAS_MAINTENANCE_NO": "",
    "SRC_INFO8": "",
    "SRC_INFO7": "",
    "SRC_INFO6": "",
    "SRC_NECLASS": "MGS-Management",
    "SRC_INFO5": "",
    "SRC_INFO4": "10.144.216.139",
    "SRC_INFO3": "贵州",
    "SRC_INFO2": "cn-gz1",
    "IAS_ORDERNO": "",
    "SRC_INFO1": "HONOR_POOL",
    "SRC_ALARMSUBTYPE": "",
    "groupID": "",
    "SRC_VENDOR": "HW",
    "SRC_CITY": "贵州",
    "IAS_FIRSTTIME": "2023-11-22 20:04:43",
    "SRC_NETTYPE": "CMP-TYY",
    "IAS_ALARMCAUSE": "计算节点系统盘/dev/sda的磁盘IO使用率持续高，请检查是否存在异常读写！",
    "IAS_IMPCATINFO": "",
    "IAS_CLEARTIME": "2023-11-22 20:15:43",
    "IAS_ALARMID": "17168398",
    "IAS_TALLY": 1,
    "IAS_IS_MAINTENANCE": 0,
    "SRC_ADDITIONALTEXT": "CloudService=MGS-Management,Service=MGS-nova-compute,MicroService=MGS-cascaded-nova-compute,NativeMeDn=10.144.216.139,regionId=cn-gz1,ItemName=diskio.util,InstanceInfo=sda3,trigger=metric=Diskio@diskio.util:value.last(0)>98,currentValue=Diskio@diskio.util:last()>98,value:97.31,mk:sda3",
    "IAS_ORDERSTATUS": 0
}`
	regFieldRaw, _ := regexp.Compile(`"(\w+)":`)
	fieldListRaw := regFieldRaw.FindAllStringSubmatch(str, -1)
	var goCode string
	for _, line := range fieldListRaw {
		if line[0] == "" {
			continue
		}
		goCode += fmt.Sprintf("%s       string `gorm:\"column:%s;comment:%s\" json:\"%s\"`\n", strings.ToUpper(line[1][:1])+line[1][1:], line[1], "", line[1])
	}
	global.Log.Info(goCode)
}

func main() {
	config.Load()
	global.InitLog("compare_field")
	global.InitDefine()

	fieldToCode()
}
