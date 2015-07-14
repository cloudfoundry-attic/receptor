package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/shared"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/storeadapter"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	"github.com/gogo/protobuf/proto"
	"github.com/pivotal-golang/lager"
)

func main() {
	logger := lager.NewLogger("perf")

	// full
	// var doraPre = `{"setup":{"serial":{"actions":[{"download":{"from":"http://file-server.service.consul:8080/v1/static/buildpack_app_lifecycle/buildpack_app_lifecycle.tgz","to":"/tmp/lifecycle","cache_key":"buildpack-cflinuxfs2-lifecycle","user":"vcap"}},{"download":{"from":"http://cloud-controller-ng.service.consul:9022/internal/v2/droplets/184aa517-b519-4e45-9c02-6bb126cbe9c5/4d260f734809cb79f65d04540a81ef64fd04a2ee/download","to":".","cache_key":"droplets-184aa517-b519-4e45-9c02-6bb126cbe9c5-fa1b700c-a58a-45b3-b1c2-3a670c4761c1","user":"vcap"}}]}},"action":{"codependent":{"actions":[{"run":{"path":"/tmp/lifecycle/launcher","args":["app","","{\"start_command\":\"bundle exec rackup config.ru -p $PORT\"}"],"env":[{"name":"VCAP_APPLICATION","value":"{\"limits\":{\"mem\":256,\"disk\":1024,\"fds\":16384},\"application_id\":\"184aa517-b519-4e45-9c02-6bb126cbe9c5\",\"application_version\":\"fa1b700c-a58a-45b3-b1c2-3a670c4761c1\",\"application_name\":\"dora\",\"version\":\"fa1b700c-a58a-45b3-b1c2-3a670c4761c1\",\"name\":\"dora\",\"space_name\":\"CATS-SPACE-1-2015_07_06-11h42m33.327s\",\"space_id\":\"84635145-9e5d-4126-a92b-2d60ac772b22\"}"},{"name":"VCAP_SERVICES","value":"{}"},{"name":"MEMORY_LIMIT","value":"256m"},{"name":"CF_STACK","value":"cflinuxfs2"},{"name":"PORT","value":"8080"}],"resource_limits":{"nofile":16384},"user":"vcap","log_source":"APP"}},{"run":{"path":"/tmp/lifecycle/diego-sshd","args":["-address=0.0.0.0:2222","-hostKey=-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDg3QiXY55Mc2cXVWfOpncyAUx6MhLDB+dGEAW7S2bwKifz/Zph\n6YDBHq0m0xjH/GXrY0X855jh+vFi1CfGGoGAWlyjY7Q2Wwynu2u2yIcE0kXK39id\nvKZLqx+GLM6hqjbJqfw2EOvGx97DXR7gO2HbJYfMTgIpUMsapMYFMyuK9wIDAQAB\nAoGBAISv6THsBqz2LA8IxoiakhtfyNESWx/aug4NxlQO2l89gPXo4ACG2QMcJvCS\nAD2CImIT4miqAPzYJzg6GH49hcwsBqO2DPy4dn+3D2h/Z0KHx6o1xthhc7uXMSFg\ng1qHeXrrqVgCGyqE3Ug6FNaiVnk/7zeYOHSGwbKyWPAGKs+BAkEA5GzVXM6xi3a1\nweBCz4Cz0j95TWVcuT49w9CWjcQ40BI0LZBK0wJa/SXTKnqr9cLv33DuZkVbja8y\nF5rmczrRvQJBAPwCIUyVvtEnNVqR6raCwWI/ZPjCLIQ6AA/lYvQPa6PjHJnxqsTU\nDFxpag8Sq5eehksC55mtQW2X37vM8djdaMMCQQCTgA+amUGea+5cHgMmWNZFKoWa\ny5w/ZgiePEArlQyWl1qoHWejr/6vPtCHuqT10oXwg8z9r0W6TOoMwhKTT+UFAkAP\nSfHLO6p/9ej+vauHtxcUZtQxY1ZgD0TBsiD2vZjCMJ0jmc3KczLsyFhu4asXX761\n/k8eu6wkgfpI4n4psgURAkBQLLDINdTVQSKhTAGjnZbocdzxdHYNra9yjBdJGClg\nSKUYPnIl8B3CGr295dJHpevBbrVBHh02UavFYgsDr/wf\n-----END RSA PRIVATE KEY-----\n","-authorizedKey=ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQDECyl47KFsOqn1oAhvW+URsqi2yluvX+RAVqDveg2RQ/jlyoEnZ/jZnQiwmSw3rW4bOBO8vNJ2I3RYjfwWVRLMSlwNnjIgX3eV3rz+Zxc62neoKeCmRl1wb7XHeGDnWaq7prlkEk3glY9VPY9p6j/YkpNQDo7pdu6TgNXkLY142w==\n","-inheritDaemonEnv","-logLevel=fatal"],"env":[{"name":"VCAP_APPLICATION","value":"{\"limits\":{\"mem\":256,\"disk\":1024,\"fds\":16384},\"application_id\":\"184aa517-b519-4e45-9c02-6bb126cbe9c5\",\"application_version\":\"fa1b700c-a58a-45b3-b1c2-3a670c4761c1\",\"application_name\":\"dora\",\"version\":\"fa1b700c-a58a-45b3-b1c2-3a670c4761c1\",\"name\":\"dora\",\"space_name\":\"CATS-SPACE-1-2015_07_06-11h42m33.327s\",\"space_id\":\"84635145-9e5d-4126-a92b-2d60ac772b22\"}"},{"name":"VCAP_SERVICES","value":"{}"},{"name":"MEMORY_LIMIT","value":"256m"},{"name":"CF_STACK","value":"cflinuxfs2"},{"name":"PORT","value":"8080"}],"resource_limits":{"nofile":16384},"user":"vcap"}}]}},"monitor":{"timeout":{"action":{"run":{"path":"/tmp/lifecycle/healthcheck","args":["-port=8080"],"env":null,"resource_limits":{"nofile":1024},"user":"vcap","log_source":"HEALTH"}},"timeout":30000000000}},"process_guid":"`
	var guidPrefix = "184aa517-b519-4e45-9c02-6bb126cbe9c5-fa1b700c-a58a-45b3-b1c2-3a670c"
	// var doraPost = `","domain":"cf-apps","rootfs":"preloaded:cflinuxfs2","instances":1,"env":[{"name":"LANG","value":"en_US.UTF-8"}],"start_timeout":60,"disk_mb":1024,"memory_mb":256,"cpu_weight":1,"privileged":true,"ports":[8080,2222],"routes":{"cf-router":[{"hostnames":["dora.10.244.0.34.xip.io"],"port":8080}],"diego-ssh":{"container_port":2222,"host_fingerprint":"50:28:aa:56:a3:03:3f:e0:19:32:03:c7:a2:f5:25:b2","private_key":"-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDECyl47KFsOqn1oAhvW+URsqi2yluvX+RAVqDveg2RQ/jlyoEn\nZ/jZnQiwmSw3rW4bOBO8vNJ2I3RYjfwWVRLMSlwNnjIgX3eV3rz+Zxc62neoKeCm\nRl1wb7XHeGDnWaq7prlkEk3glY9VPY9p6j/YkpNQDo7pdu6TgNXkLY142wIDAQAB\nAoGBALkg0UkgLE/IFjedqFmArhDIZgo3jd1O8HzRUajT2XwUdDaLxOsxhA37/PjH\nrLnnTNLnYbwZk6V8VaJKcoOkUtpu+BWEVP26eIlnKk/fqQcGMklphqnKhAkzohwj\n3vAjKaVzvwfmEJm79Ctmh41iHheTU4/s10+7+JdcOlfxAKgBAkEA+JGMjTcTfoa6\nEaQPl9SdlMxklQToQoI3i8Yd8av9yYfWUH9E23YerfX0B96X5LcApAfJmaoBvQln\nbzRFJF6UCwJBAMnnm0pAqty+zrKssVl7X2SrupkHFD9/RvSzPLHhCmGzZ/62kOW7\nbnX4QdxDiMgzXBh1q8hjdpqfM6j8vU0eYHECQA3LXgZ0OQO7jFXwSeE+LmSUlzxh\n4lXWjiiWnRDNX68wd6dN+M9JFdjHnnxVUQ6jTUjNGdYKRkBsZi4Ys4GaMhMCQQCZ\njhcRwtrv5gIn66U6K9ViKCVTSwoAPNmHM2Ye1sthgOO/2bObtRAOko/saER4Fm+d\nfqj2T4cdk6TjicyjAU5RAkBnH3rVeZTfiLluRYtECbheKzehuCKhFgQwOTW4upXd\nCtQravNMn86Dsvztz7daSvnziqHvPSHCPixxMwDlyd9E\n-----END RSA PRIVATE KEY-----\n"}},"log_guid":"184aa517-b519-4e45-9c02-6bb126cbe9c5","log_source":"CELL","metrics_guid":"184aa517-b519-4e45-9c02-6bb126cbe9c5","annotation":"1436275647.3182652","egress_rules":[{"protocol":"tcp","destinations":["0.0.0.0/0"],"ports":[53],"log":false},{"protocol":"udp","destinations":["0.0.0.0/0"],"ports":[53],"log":false},{"protocol":"all","destinations":["0.0.0.0-9.255.255.255"],"log":false},{"protocol":"all","destinations":["11.0.0.0-169.253.255.255"],"log":false},{"protocol":"all","destinations":["169.255.0.0-172.15.255.255"],"log":false},{"protocol":"all","destinations":["172.32.0.0-192.167.255.255"],"log":false},{"protocol":"all","destinations":["192.169.0.0-255.255.255.255"],"log":false}],"modification_tag":{"epoch":"97395e67-5a9b-4562-6965-43453388f371","index":1}}`

	// 1 small actions
	//	var doraPre = `{"action":{"run":{"path":"/tmp/lifecycle/launcher"}},"process_guid":"`
	//	var guidPrefix = "184aa517-b519-4e45-9c02-6bb126cbe9c5-fa1b700c-a58a-45b3-b1c2-3a670c"
	//	var doraPost = `","domain":"cf-apps","rootfs":"preloaded:cflinuxfs2","instances":1,"start_timeout":60,"disk_mb":1024,"memory_mb":256,"cpu_weight":1,"privileged":true,"ports":[8080],"routes":{"cf-router":[{"hostnames":["dora.10.244.0.34.xip.io"],"port":8080}]},"log_guid":"184aa517-b519-4e45-9c02-6bb126cbe9c5","log_source":"CELL","metrics_guid":"184aa517-b519-4e45-9c02-6bb126cbe9c5","annotation":"1436275647.3182652","modification_tag":{"epoch":"97395e67-5a9b-4562-6965-43453388f371","index":1}}`

	var u64 uint64 = 10
	cfrouter := json.RawMessage([]byte(`[{"hostnames":["dora.10.244.0.34.xip.io"],"port":8080}]`))
	diegossh := json.RawMessage([]byte(`{"container_port": 2222, "host_fingerprint": "50:28:aa:56:a3:03:3f:e0:19:32:03:c7:a2:f5:25:b2", "private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDECyl47KFsOqn1oAhvW+URsqi2yluvX+RAVqDveg2RQ/jlyoEn\nZ/jZnQiwmSw3rW4bOBO8vNJ2I3RYjfwWVRLMSlwNnjIgX3eV3rz+Zxc62neoKeCm\nRl1wb7XHeGDnWaq7prlkEk3glY9VPY9p6j/YkpNQDo7pdu6TgNXkLY142wIDAQAB\nAoGBALkg0UkgLE/IFjedqFmArhDIZgo3jd1O8HzRUajT2XwUdDaLxOsxhA37/PjH\nrLnnTNLnYbwZk6V8VaJKcoOkUtpu+BWEVP26eIlnKk/fqQcGMklphqnKhAkzohwj\n3vAjKaVzvwfmEJm79Ctmh41iHheTU4/s10+7+JdcOlfxAKgBAkEA+JGMjTcTfoa6\nEaQPl9SdlMxklQToQoI3i8Yd8av9yYfWUH9E23YerfX0B96X5LcApAfJmaoBvQln\nbzRFJF6UCwJBAMnnm0pAqty+zrKssVl7X2SrupkHFD9/RvSzPLHhCmGzZ/62kOW7\nbnX4QdxDiMgzXBh1q8hjdpqfM6j8vU0eYHECQA3LXgZ0OQO7jFXwSeE+LmSUlzxh\n4lXWjiiWnRDNX68wd6dN+M9JFdjHnnxVUQ6jTUjNGdYKRkBsZi4Ys4GaMhMCQQCZ\njhcRwtrv5gIn66U6K9ViKCVTSwoAPNmHM2Ye1sthgOO/2bObtRAOko/saER4Fm+d\nfqj2T4cdk6TjicyjAU5RAkBnH3rVeZTfiLluRYtECbheKzehuCKhFgQwOTW4upXd\nCtQravNMn86Dsvztz7daSvnziqHvPSHCPixxMwDlyd9E\n-----END RSA PRIVATE KEY-----\n"}`))
	r := models.DesiredLRP{
		ProcessGuid: proto.String("184aa517-b519-4e45-9c02-6bb126cbe9c5-fa1b700c-a58a-45b3-b1c2-3a670c"),
		Domain:      proto.String("cf-apps"),
		RootFs:      proto.String("preloaded:cflinuxfs2"),
		Instances:   proto.Int(1),
		EnvironmentVariables: []*models.EnvironmentVariable{
			&models.EnvironmentVariable{Name: proto.String("LANG"), Value: proto.String("en_US.UTF-8")},
		},
		Setup: models.WrapAction(&models.SerialAction{
			Actions: []*models.Action{
				models.WrapAction(&models.DownloadAction{
					From:     proto.String("http://file-server.service.consul:8080/v1/static/buildpack_app_lifecycle/buildpack_app_lifecycle.tgz"),
					To:       proto.String("/tmp/lifecycle"),
					CacheKey: proto.String("buildpack-cflinuxfs2-lifecycle"),
					User:     proto.String("vcap"),
				}),
				models.WrapAction(&models.DownloadAction{
					From:     proto.String("http://cloud-controller-ng.service.consul:9022/internal/v2/droplets/184aa517-b519-4e45-9c02-6bb126cbe9c5/4d260f734809cb79f65d04540a81ef64fd04a2ee/download"),
					To:       proto.String("."),
					CacheKey: proto.String("droplets-184aa517-b519-4e45-9c02-6bb126cbe9c5-fa1b700c-a58a-45b3-b1c2-3a670c4761c1"),
					User:     proto.String("vcap"),
				}),
			},
		}),

		Action: models.WrapAction(&models.CodependentAction{
			Actions: []*models.Action{
				models.WrapAction(&models.RunAction{
					Path: proto.String("/tmp/lifecycle/launcher"),
					Args: []string{"app", "", "{\"start_command\":\"bundle exec rackup config.ru -p $PORT\"}"},
					Env: []*models.EnvironmentVariable{
						&models.EnvironmentVariable{
							Name:  proto.String("VCAP_APPLICATION"),
							Value: proto.String("{\"limits\":{\"mem\":256,\"disk\":1024,\"fds\":16384},\"application_id\":\"184aa517-b519-4e45-9c02-6bb126cbe9c5\",\"application_version\":\"fa1b700c-a58a-45b3-b1c2-3a670c4761c1\",\"application_name\":\"dora\",\"version\":\"fa1b700c-a58a-45b3-b1c2-3a670c4761c1\",\"name\":\"dora\",\"space_name\":\"CATS-SPACE-1-2015_07_06-11h42m33.327s\",\"space_id\":\"84635145-9e5d-4126-a92b-2d60ac772b22\"}"),
						},
						&models.EnvironmentVariable{
							Name:  proto.String("VCAP_SERVICES"),
							Value: proto.String("{}"),
						},
						&models.EnvironmentVariable{
							Name:  proto.String("MEMORY_LIMIT"),
							Value: proto.String("256m"),
						},
						&models.EnvironmentVariable{
							Name:  proto.String("CF_STACK"),
							Value: proto.String("cflinuxfs2"),
						},
						&models.EnvironmentVariable{
							Name:  proto.String("PORT"),
							Value: proto.String("8080"),
						},
					},
					ResourceLimits: &models.ResourceLimits{
						Nofile: &u64,
					},
					User:      proto.String("vcap"),
					LogSource: proto.String("APP"),
				}),
				models.WrapAction(&models.RunAction{
					Path: proto.String("/tmp/lifecycle/diego-sshd"),
					Args: []string{"-address=0.0.0.0:2222", "-hostKey=-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDg3QiXY55Mc2cXVWfOpncyAUx6MhLDB+dGEAW7S2bwKifz/Zph\n6YDBHq0m0xjH/GXrY0X855jh+vFi1CfGGoGAWlyjY7Q2Wwynu2u2yIcE0kXK39id\nvKZLqx+GLM6hqjbJqfw2EOvGx97DXR7gO2HbJYfMTgIpUMsapMYFMyuK9wIDAQAB\nAoGBAISv6THsBqz2LA8IxoiakhtfyNESWx/aug4NxlQO2l89gPXo4ACG2QMcJvCS\nAD2CImIT4miqAPzYJzg6GH49hcwsBqO2DPy4dn+3D2h/Z0KHx6o1xthhc7uXMSFg\ng1qHeXrrqVgCGyqE3Ug6FNaiVnk/7zeYOHSGwbKyWPAGKs+BAkEA5GzVXM6xi3a1\nweBCz4Cz0j95TWVcuT49w9CWjcQ40BI0LZBK0wJa/SXTKnqr9cLv33DuZkVbja8y\nF5rmczrRvQJBAPwCIUyVvtEnNVqR6raCwWI/ZPjCLIQ6AA/lYvQPa6PjHJnxqsTU\nDFxpag8Sq5eehksC55mtQW2X37vM8djdaMMCQQCTgA+amUGea+5cHgMmWNZFKoWa\ny5w/ZgiePEArlQyWl1qoHWejr/6vPtCHuqT10oXwg8z9r0W6TOoMwhKTT+UFAkAP\nSfHLO6p/9ej+vauHtxcUZtQxY1ZgD0TBsiD2vZjCMJ0jmc3KczLsyFhu4asXX761\n/k8eu6wkgfpI4n4psgURAkBQLLDINdTVQSKhTAGjnZbocdzxdHYNra9yjBdJGClg\nSKUYPnIl8B3CGr295dJHpevBbrVBHh02UavFYgsDr/wf\n-----END RSA PRIVATE KEY-----\n", "-authorizedKey=ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQDECyl47KFsOqn1oAhvW+URsqi2yluvX+RAVqDveg2RQ/jlyoEnZ/jZnQiwmSw3rW4bOBO8vNJ2I3RYjfwWVRLMSlwNnjIgX3eV3rz+Zxc62neoKeCmRl1wb7XHeGDnWaq7prlkEk3glY9VPY9p6j/YkpNQDo7pdu6TgNXkLY142w==\n", "-inheritDaemonEnv", "-logLevel=fatal"},
					Env: []*models.EnvironmentVariable{
						&models.EnvironmentVariable{
							Name:  proto.String("VCAP_APPLICATION"),
							Value: proto.String("{\"limits\":{\"mem\":256,\"disk\":1024,\"fds\":16384},\"application_id\":\"184aa517-b519-4e45-9c02-6bb126cbe9c5\",\"application_version\":\"fa1b700c-a58a-45b3-b1c2-3a670c4761c1\",\"application_name\":\"dora\",\"version\":\"fa1b700c-a58a-45b3-b1c2-3a670c4761c1\",\"name\":\"dora\",\"space_name\":\"CATS-SPACE-1-2015_07_06-11h42m33.327s\",\"space_id\":\"84635145-9e5d-4126-a92b-2d60ac772b22\"}"),
						},
						&models.EnvironmentVariable{
							Name:  proto.String("VCAP_SERVICES"),
							Value: proto.String("{}"),
						},
						&models.EnvironmentVariable{
							Name:  proto.String("MEMORY_LIMIT"),
							Value: proto.String("256m"),
						},
						&models.EnvironmentVariable{
							Name:  proto.String("CF_STACK"),
							Value: proto.String("cflinuxfs2"),
						},
						&models.EnvironmentVariable{
							Name:  proto.String("PORT"),
							Value: proto.String("8080"),
						},
					},
					ResourceLimits: &models.ResourceLimits{
						Nofile: &u64,
					},
					User: proto.String("vcap"),
				}),
			},
		}),
		Monitor: models.WrapAction(&models.TimeoutAction{
			Action: models.WrapAction(&models.RunAction{
				Path: proto.String("/tmp/lifecycle/healthcheck"),
				Args: []string{"-port=8080"},
				Env:  nil,
				ResourceLimits: &models.ResourceLimits{
					Nofile: &u64,
				},
				User:      proto.String("vcap"),
				LogSource: proto.String("HEALTH"),
			}),
			Timeout: proto.Int64(3000000),
		}),
		StartTimeout: proto.Uint32(60),
		DiskMb:       proto.Int32(1024),
		MemoryMb:     proto.Int32(256),
		CpuWeight:    proto.Uint32(1),
		Privileged:   proto.Bool(true),
		Ports:        []uint32{8080, 2222},
		Routes: &models.Routes{
			"cf-router": &cfrouter,
			"diego-ssh": &diegossh,
		},
		LogGuid:     proto.String("184aa517-b519-4e45-9c02-6bb126cbe9c5"),
		LogSource:   proto.String("CELL"),
		MetricsGuid: proto.String("184aa517-b519-4e45-9c02-6bb126cbe9c5"),
		Annotation:  proto.String("1436275647.3182652"),
		EgressRules: []*models.SecurityGroupRule{
			{
				Protocol:     proto.String("tcp"),
				Destinations: []string{"0.0.0.0/0"},
				Ports:        []uint32{53},
				Log:          proto.Bool(false),
			},
			{
				Protocol:     proto.String("udp"),
				Destinations: []string{"0.0.0.0/0"},
				Ports:        []uint32{53},
				Log:          proto.Bool(false),
			},
			{
				Protocol:     proto.String("all"),
				Destinations: []string{"0.0.0.0-9.255.255.255"},
				Log:          proto.Bool(false),
			},
			{
				Protocol:     proto.String("all"),
				Destinations: []string{"11.0.0.0-169.253.255.255"},
				Log:          proto.Bool(false),
			},
			{
				Protocol:     proto.String("all"),
				Destinations: []string{"169.255.0.0-172.15.255.255"},
				Log:          proto.Bool(false),
			},
			{
				Protocol:     proto.String("all"),
				Destinations: []string{"172.32.0.0-192.167.255.255"},
				Log:          proto.Bool(false),
			},
			{
				Protocol:     proto.String("all"),
				Destinations: []string{"192.169.0.0-255.255.255.255"},
				Log:          proto.Bool(false),
			},
		},
	}

	immutable := r
	immutable.Routes = nil
	immutable.Instances = nil
	immutable.Annotation = nil
	immutable.Ports = nil

	mutable := r
	mutable.Domain = nil
	mutable.RootFs = nil
	mutable.EnvironmentVariables = nil
	mutable.Setup = nil
	mutable.Action = nil
	mutable.Monitor = nil
	immutable.StartTimeout = nil
	immutable.DiskMb = nil
	immutable.MemoryMb = nil
	immutable.CpuWeight = nil
	immutable.Privileged = nil
	mutable.EgressRules = nil
	mutable.LogGuid = nil
	mutable.LogSource = nil
	mutable.MetricsGuid = nil
	mutable.EgressRules = nil

	var count = 100000

	workPool, err := workpool.NewWorkPool(10)
	if err != nil {
		logger.Fatal("failed-to-construct-etcd-adapter-workpool", err, lager.Data{"num-workers": 100}) // should never happen
	}

	// NOTE: Modify your /etc/hosts -- 10.244.16.130 database-z1-0.etcd.service.consul
	options := &etcdstoreadapter.ETCDOptions{
		CertFile:    "/Users/fraenkel/workspace/diego-release/manifest-generation/bosh-lite-stubs/etcd-certs/client.crt",
		KeyFile:     "/Users/fraenkel/workspace/diego-release/manifest-generation/bosh-lite-stubs/etcd-certs/client.key",
		CAFile:      "/Users/fraenkel/workspace/diego-release/manifest-generation/bosh-lite-stubs/etcd-certs/diego-ca.crt",
		IsSSL:       true,
		ClusterUrls: []string{"https://database-z1-0.etcd.service.consul:4001"},
	}

	etcdAdapter, err := etcdstoreadapter.New(options, workPool)
	if err != nil {
		logger.Fatal("adapter", err)
	}

	start := time.Now()
	for i := 0; i < count; i++ {
		mutable.ProcessGuid = proto.String(guidPrefix + strconv.Itoa(i))
		m, err := proto.Marshal(&mutable)
		if err != nil {
			logger.Fatal("mutable", err)
		}

		err = etcdAdapter.Create(storeadapter.StoreNode{
			Key:   shared.DesiredLRPSchemaPathByProcessGuid(mutable.GetProcessGuid()),
			Value: m,
		})
	}
	end := time.Now()
	println("insert", end.Sub(start).String())
}
