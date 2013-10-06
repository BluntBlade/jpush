package jpush

import (
    "bytes"
    "fmt"
    "strings"
    "crypto/md5"
    "encoding/json"
    "net/http"
    "net/url"
)

type Pusher struct {
    host            string
    appKey          string
    masterSecret    string
} // Pusher

func NewPusher(host, appKey, masterSecret string) (push *Pusher) {
    push = &Pusher{
        host:           host,
        appKey:         appKey,
        masterSecret:   masterSecret,
    }
    return
} // NewPusher

// 接收者类型
const (
    RECV_BY_TAG   = "2"  // 由tag列表标示的终端群组接收该消息（群发），使用逗号分隔
    RECV_BY_ALIAS = "3"  // 由alias列表标示的终端接收该消息（一对多），使用逗号分隔
    RECV_BY_APP   = "4"  // 由app标示的所有终端接收该消息（应用范围广播），不需要填接收者
)

// 消息类型
const (
    MSG_NOTIFICATION = "1"
    MSG_USERDEFINED  = "2" // 只支持Android终端
)

type Push struct {
    SendNo          uint32
    ReceiverType    string
    ReceiverValue   []string
    MsgType         string
    MsgContent      string
    SendDescription string      // 关于本消息的描述，不会发送给接收者
    Platform        []string    // 端手机的平台类型，如：android，ios，使用逗号分隔
    TimeToLive      uint        // 离线保存时间，单位：秒，最长为10天（864000秒）
                                // 0 表示该消息不保存离线，即：用户在线马上发出，当前不在线用户将不会收到此消息
    OverrideMsgId   string      // 待覆盖的上一条消息的ID
} // Push

type PushRet struct {
    ErrCode         uint   `json:"errcode"`
    ErrMsg          string `json:"errmsg"`
    MsgID           string `json:"msg_id"`
} // PushRet

func (pr *Pusher) Push(ret *PushRet, p *Push) (err error) {
    kvs := url.Values{}

    kvs.Add("sendno", fmt.Sprintf("%d", p.SendNo))
    kvs.Add("app_key", pr.appKey)
    kvs.Add("receiver_type", p.ReceiverType)

    receiverValue := ""
    if p.ReceiverType == RECV_BY_TAG || p.ReceiverType == RECV_BY_ALIAS {
        receiverValue = strings.Join(p.ReceiverValue, ",")
        kvs.Add("receiver_value", receiverValue)
    }

    {
        msg := fmt.Sprintf("%d", p.SendNo) + p.ReceiverType + receiverValue + pr.masterSecret

        var tmp [16]byte
        md5er := md5.New()
        _, err = md5er.Write([]byte(msg))
        if err != nil {
            return
        }

        md5Val := md5er.Sum(tmp[:0])
        kvs.Add("verification_code", fmt.Sprintf("%x", md5Val[:]))
    }

    kvs.Add("msg_type", p.MsgType)
    kvs.Add("msg_content", p.MsgContent)

    if p.SendDescription != "" {
        kvs.Add("send_description", p.SendDescription)
    }

    kvs.Add("platform", strings.Join(p.Platform, ","))
    kvs.Add("time_to_live", fmt.Sprintf("%d", p.TimeToLive))

    if p.OverrideMsgId != "" {
        kvs.Add("override_msg_id", p.OverrideMsgId)
    }

    body := kvs.Encode()
    resp, err := http.Post(pr.host, "application/x-www-form-urlencoded", strings.NewReader(body))
    if err != nil {
        return
    }
    defer resp.Body.Close()

    if ret != nil {
        jsoner := json.NewDecoder(resp.Body)
        err = jsoner.Decode(ret)
    }
    return
} // Push

type Notification struct {
    NBuilderId uint                   `json:"n_builder_id"` // 可选
    NTitle     string                 `json:"n_title"`      // 可选
    NContent   string                 `json:"n_content"`
    NExtras    map[string]interface{} `json:"n_extras"`     // 可选
} // Notification

func (noti *Notification) MarshalJSON() ([]byte, error) {
    tmp := map[string]interface{}{}

    tmp["n_builder_id"] = noti.NBuilderId
    tmp["n_content"] = noti.NContent

    if noti.NTitle != "" {
        tmp["n_title"] = noti.NTitle
    }
    if noti.NExtras != nil {
        tmp["n_extras"] = noti.NExtras
    }

    buf := new(bytes.Buffer)
    err := json.NewEncoder(buf).Encode(&tmp)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
} // MarshalJSON

type UserDefinedMessage struct {
    ContentType interface{}             `json:"content_type"`
    Title       string                  `json:"title"`
    Message     string                  `json:"message"`
    Extras      map[string]interface{}  `json:"extras"`
} // UserDefinedMessage

func (msg *UserDefinedMessage) MarshalJSON() ([]byte, error) {
    tmp := map[string]interface{}{}

    tmp["message"] = msg.Message

    if msg.ContentType != nil {
        tmp["content_type"] = msg.ContentType
    }
    if msg.Title != "" {
        tmp["title"] = msg.Title
    }
    if msg.Extras != nil {
        tmp["extras"] = msg.Extras
    }

    buf := new(bytes.Buffer)
    err := json.NewEncoder(buf).Encode(&tmp)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
} // MarshalJSON
