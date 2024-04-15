package hdfs_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/colinmarc/hdfs/v2"
	"github.com/colinmarc/hdfs/v2/hadoopconf"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
)

func Test_Hdfs_(t *testing.T) {
	addr := []string{"172.20.27.145:9000"}
	user := "hdfsuser"

	client, err := getConnectKerberos(addr, user)
	if err != nil {
		return
	}
	map1, err := client.ListXAttrs("/tmp/zml")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(map1)
	// client, err := getConnect(addr, user)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	fmt.Println(client.Name())
	fmt.Println("over!")
}
func getConnectKerberos(addr []string, username string) (*hdfs.Client, error) {

	//*1.填写hdfs配置
	// options := hdfs.ClientOptions{
	// 	Addresses: addr,
	// 	User:      username,
	// }
	//*2.从xml导入hdfs配置
	hcf, err := hadoopconf.Load("./conf")
	if err != nil {
		fmt.Println(err)
	}
	options := hdfs.ClientOptionsFromConf(hcf)

	//*导入kerberos配置
	var (
		ktp                  = "./conf/hdfs.keytab"
		cfp                  = "./conf/krb5.conf"
		servicePrincipleName = "hdfs/sjzt-wuhan-9"
		uname                = "hdfs/sjzt-wuhan-9"
		realm                = "HADOOP.COM"
	)

	kt, err := keytab.Load(ktp)
	if err != nil {
		fmt.Printf("keytab.Load err: %v\n", err)
	}
	cf, err := config.Load(cfp)
	if err != nil {
		fmt.Printf("config.Load err: %v\n", err)
	}
	options.KerberosServicePrincipleName = servicePrincipleName
	options.KerberosClient = client.NewWithKeytab(uname, realm, kt, cf)

	// isconf, err := options.KerberosClient.IsConfigured()
	// if err != nil {
	// 	fmt.Printf("IsConfigured err: %v\n", err)
	// }
	// fmt.Printf("IsConfigured: %v\n", isconf)
	//*测试登录
	// options.KerberosClient.Login()
	// options.KerberosClient.AffirmLogin()
	mt1, tek1, ok1 := options.KerberosClient.GetCachedTicket(options.KerberosServicePrincipleName)
	p1, _ := json.Marshal(mt1)
	fmt.Printf("GetCachedTicket:mt1:%s\n, tek1:%v, ok1:%v\n", p1, tek1, ok1)
	mt2, tek2, ok2 := options.KerberosClient.GetServiceTicket(options.KerberosServicePrincipleName)
	p2, _ := json.Marshal(mt2)
	fmt.Printf("GetServiceTicket:mt2:%s\n, tek2:%v, ok2:%v\n", p2, tek2, ok2)

	//*测试
	hdfsclient, err := hdfs.NewClient(options)

	if err != nil {
		fmt.Println(err)
	}
	return hdfsclient, nil
}

func getConnect(addr []string, username string) (*hdfs.Client, error) {
	conf := hdfs.ClientOptions{
		Addresses: addr,
		User:      username,
	}

	client, err := hdfs.NewClient(conf)

	if err != nil {
		return nil, err
	}
	return client, nil
}
