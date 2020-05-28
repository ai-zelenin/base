package pharvester

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sparrc/go-ping"
	"golang.org/x/net/proxy"
	"h12.io/socks"
)

type Proxy struct {
	ID                int64         `json:"id" yaml:"id" gorm:"PRIMARY_KEY; AUTO_INCREMENT"`
	Host              string        `json:"host" yaml:"host" gorm:"column:host; unique_index:address;"`
	Port              string        `json:"port" yaml:"port" gorm:"column:port; unique_index:address;"`
	Protocol          string        `json:"protocol" yaml:"protocol" gorm:"column:protocol; index:protocol;"`
	Country           string        `json:"country" yaml:"country" gorm:"column:country; index:country;"`
	Source            string        `json:"source" yaml:"source" gorm:"column:source;"`
	Score             int64         `json:"score" yaml:"score" gorm:"column:score;"`
	Status            int64         `json:"status" yaml:"status" gorm:"column:status;"`
	Username          string        `json:"username" yaml:"username" gorm:"column:username;"`
	Password          string        `json:"password" yaml:"password" gorm:"column:password;"`
	LastCheck         time.Time     `json:"last_check" yaml:"last_check" gorm:"column:last_check;"`
	CheckNumber       int           `json:"check_number" gorm:"column:check_number;"`
	AvgPingTiming     time.Duration `json:"avg_ping_timing" yaml:"avg_ping_timing" gorm:"column:avg_ping_Timing;"`
	AvgRequestTiming  time.Duration `json:"avg_request_timing" yaml:"avg_request_timing" gorm:"column:avg_request_Timing;"`
	AvgResponseTiming time.Duration `json:"avg_response_timing" yaml:"avg_response_timing" gorm:"column:avg_response_Timing;"`
	CreatedAt         time.Time     `json:"created_at" yaml:"created_at" gorm:"column:created_at;"`
	Client            *http.Client  `json:"-" yaml:"-" gorm:"-"`
	Metrics           *ProxyMetrics `json:"-" yaml:"-" gorm:"-"`
	mx                sync.RWMutex  `json:"-" yaml:"-" gorm:"-"`
}

func (p *Proxy) URL() string {
	p.mx.RLock()
	defer p.mx.RUnlock()
	return fmt.Sprintf("%s://%s:%s", strings.ToLower(p.Protocol), p.Host, p.Port)
}

func (p *Proxy) InitClient(timeout time.Duration) error {
	u, err := url.Parse(p.URL())
	if err != nil {
		return err
	}
	if u.Host == "" || u.Port() == "" || u.Scheme == "" {
		return NewError(ErrCodeInvalidProxyURL, nil, "insufficient proxy data")
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	switch u.Scheme {
	case "socks5":
		var auth *proxy.Auth
		if len(p.Username) > 0 {
			auth = &proxy.Auth{
				User:     p.Username,
				Password: p.Password,
			}
		}
		dialer, err := proxy.SOCKS5("tcp", net.JoinHostPort(p.Host, p.Port), auth, proxy.Direct)
		if err != nil {
			return NewError(ErrCodeCannotCreateProxyDialer, err)
		}
		transport.Dial = dialer.Dial
	case "socks4":
		transport.Dial = socks.DialSocksProxy(socks.SOCKS4, net.JoinHostPort(p.Host, p.Port))
	default:
		transport.Proxy = http.ProxyURL(u)

	}
	p.Client = &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return nil
}
func (p *Proxy) CheckProxy(target string) error {
	start := time.Now()
	resp, err := p.Client.Get(target)
	if err != nil {
		return NewError(ErrCodeRequestFail, err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NewError(ErrCodeRequestFail, err)
	}
	parts := strings.Split(string(data), " ")
	if len(parts) != 3 {
		return NewError(ErrCodeBadResponse, nil, "invalid response format from server")
	}
	msec, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return NewError(ErrCodeBadResponse, err)
	}
	nsec := int64(msec * 1e9)
	end := time.Now()

	timeToResponse := end.Sub(start)
	timeToServer := time.Unix(0, nsec).Sub(start)

	p.Metrics.AddRequestMetric(RequestMetrics{
		RequestTiming:  timeToServer,
		ResponseTiming: timeToResponse,
	})
	return nil
}

func (p *Proxy) CheckPing() error {
	pinger, err := ping.NewPinger(p.Host)
	if err != nil {
		return NewError(ErrCodeUntyped, err)
	}
	pinger.Count = 3
	pinger.Timeout = time.Second * 5

	pinger.Run()
	stats := pinger.Statistics()
	if stats.PacketLoss > 0 {
		return NewError(ErrCodePacketLoss, fmt.Errorf("route to proxy has packet loss %f", stats.PacketLoss))
	}
	p.Metrics.PingTiming = stats.AvgRtt
	return nil
}

func (p *Proxy) SetPort(port string) {
	p.mx.Lock()
	p.Port = port
	p.mx.Unlock()
}

func (p *Proxy) SetProtocol(protocol string) {
	p.mx.Lock()
	p.Protocol = protocol
	p.mx.Unlock()
}

func (p *Proxy) Marshal() ([]byte, error) {
	p.mx.RLock()
	defer p.mx.RUnlock()
	return json.MarshalIndent(p, "\t", "")
}

func (p *Proxy) String() string {
	return fmt.Sprintf("Host:%s Port:%s Protocol:%s CreatedAt:%s", p.Host, p.Port, p.Protocol, p.CreatedAt.Format(time.RFC3339))
}
