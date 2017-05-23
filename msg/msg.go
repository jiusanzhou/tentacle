package msg

import (
	"encoding/json"
	"reflect"
)

var TypeMap map[string]reflect.Type

func init() {
	TypeMap = make(map[string]reflect.Type)

	t := func(obj interface{}) reflect.Type { return reflect.TypeOf(obj).Elem() }

	TypeMap["Auth"] = t((*Auth)(nil))
	TypeMap["AuthResp"] = t((*AuthResp)(nil))
	TypeMap["Ping"] = t((*Ping)(nil))
	TypeMap["Pong"] = t((*Pong)(nil))
	TypeMap["Dial"] = t((*Dial)(nil))
	TypeMap["DialResp"] = t((*DialResp)(nil))
	TypeMap["Cmd"] = t((*Cmd)(nil))
	TypeMap["CmdResp"] = t((*CmdResp)(nil))
	TypeMap["RegTun"] = t((*RegTun)(nil))
}

type Message interface{}

type PackageMsg struct {
	MsgType [1]byte
}

type Envelope struct {
	Type    string
	Payload json.RawMessage
}

// When a client opens a new control channel to the server
// it must start by sending an Auth message.
type Auth struct {
	Version   string // protocol version
	MmVersion string // major/minor software version (informational only)
	User      string
	Password  string
	OS        string
	Arch      string

	// host name
	// self ip
	// or something else

	ClientId string // empty for new sessions
}

// A server responds to an Auth message with an
// AuthResp message over the control channel.
//
// If Error is not the empty string
// the server has indicated it will not accept
// the new session and will close the connection.
//
// The server response includes a unique ClientId
// that is used to associate and authenticate future
// proxy connections via the same field in RegProxy messages.
type AuthResp struct {
	Version    string
	MmVersion  string
	ClientId   string
	Error      string
	TunnelPort string
}

type Cmd struct {
	ClientId string
	Commands []string
}

type CmdResp struct {
	ClientId string
	Stdout   []string
}

type Dial struct {
	ClientId string
	ReqId    string
	RawAddr  []byte
	Addr     string
	Data     []byte
}

type DialResp struct {
	ClientId string
	ReqId    string
}

type RegTun struct {
	ClientId string
	ReqId    string
}

// A client or server may send this message periodically over
// the control channel to request that the remote side acknowledge
// its connection is still alive. The remote side must respond with a Pong.
type Ping struct{}

// Sent by a client or server over the control channel to indicate
// it received a Ping.
type Pong struct{}

/* -----------------------------------------------------------------------------------

██████╗ ███████╗███╗   ███╗ ██████╗ ██╗   ██╗███████╗
██╔══██╗██╔════╝████╗ ████║██╔═══██╗██║   ██║██╔════╝
██████╔╝█████╗  ██╔████╔██║██║   ██║██║   ██║█████╗
██╔══██╗██╔══╝  ██║╚██╔╝██║██║   ██║╚██╗ ██╔╝██╔══╝
██║  ██║███████╗██║ ╚═╝ ██║╚██████╔╝ ╚████╔╝ ███████╗
╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝   ╚═══╝  ╚══════╝

                              .=""--._
                 __..._    ,="_`/.--""
            ..-""__...._"""       `^"\
          .'  ,/_,.__.- _,       _  .`.
        .'       _.' .-';       /_\ \o|_
      .'       -" .-'  /        `o' /   \,-
      `"""""----""    (        `.--'`---'='
                       `..     .'.`-..-/`\
                          `";`7 'j`"--'
                          _.| |  |
                       .-'    ;  `.
                    .-'  .-   :`   ;
                 .-'_.._7___ _7   ;|.---.
                (           `"\  /--..r=`)
                 \__..--"7'`. ,`7     `}\'
       __.    .-"       /    J/}/
   .-""   \.-"        .'     `;
  :      .'         .'       ;
  ;     /          :         |       .-._
  `.   :           |         ;-.    /`. `/
    `--|           ;        /   \  ;`. ` :
      _;           ;      .'     : :  .-':
    .' \          ;  _.--'       :/ .'  /
    |  ,__     .-'"""""--.       7 /   /
    :    \`""""           `.    ' /   /
     :    J__..._           `.   ;  .'
      \    -. `-.\            `,J.-'
       `._   `._.'                   Zoe
----------------------------------------------------------------------------------- */
