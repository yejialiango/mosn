/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package healthcheck

import (
	"encoding/json"
	"mosn.io/api"
	"net"
	"reflect"
	"time"

	"mosn.io/mosn/pkg/log"
	"mosn.io/mosn/pkg/types"
)

type TCPDialSessionFactory struct{}

const (
	TCPCheckConfigKey = "tcp_check_config"
)

var tcpDefaultTimeout = api.DurationConfig{time.Second * 5}

type TCPCheckConfig struct {
	Timeout api.DurationConfig `json:"timeout,omitempty"`
}

func (f *TCPDialSessionFactory) NewSession(cfg map[string]interface{}, host types.Host) types.HealthCheckSession {
	tcpDialSession := &TCPDialSession{
		addr:    host.AddressString(),
		timeout: tcpDefaultTimeout.Duration,
	}
	v, ok := cfg[TCPCheckConfigKey]
	if !ok {
		return tcpDialSession
	}
	tcpCfg, ok := v.(*TCPCheckConfig)
	if !ok {
		tcpCfgBytes, err := json.Marshal(v)
		if err != nil {
			log.DefaultLogger.Errorf("[upstream] [health check] [tcpdial session] tcpCfg covert %+v error %+v %+v", reflect.TypeOf(v), v, err)
			return nil
		}
		tcpCfg = &TCPCheckConfig{}
		if err := json.Unmarshal(tcpCfgBytes, tcpCfg); err != nil {
			log.DefaultLogger.Errorf("[upstream] [health check] [tcpdial session] tcpCfg Unmarshal %+v error %+v %+v", reflect.TypeOf(v), v, err)
			return nil
		}
	}
	if tcpCfg.Timeout.Duration > 0 {
		tcpDialSession.timeout = tcpCfg.Timeout.Duration
	}
	return tcpDialSession
}

type TCPDialSession struct {
	addr    string
	timeout time.Duration
}

func (s *TCPDialSession) CheckHealth() bool {
	// default dial timeout, maybe already timeout by checker
	conn, err := net.DialTimeout("tcp", s.addr, s.timeout)
	if err != nil {
		if log.DefaultLogger.GetLogLevel() >= log.INFO {
			log.DefaultLogger.Infof("[upstream] [health check] [tcpdial session] dial tcp for host %s error: %v", s.addr, err)
		}
		return false
	}
	conn.Close()
	return true
}

func (s *TCPDialSession) OnTimeout() {}
