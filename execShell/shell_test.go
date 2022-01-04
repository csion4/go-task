package execShell

//func main() {
//	//// 在宿主机中执行shell脚本
//	//cmd := exec.Command("cmd", "/C", "dir", "set aa=qwe123", "echo %aa%")
//	//cmd.Dir = "D:/"
//	//
//	//output, err := cmd.CombinedOutput()
//	//if err != nil {
//	//	panic(err.Error())
//	//}
//	//fmt.Println(string(output))
//
//	c := exec.Command("bash", "-c", "ping www.baidu.com")  // mac or linux
//	stdout, err := c.StdoutPipe()
//	if err != nil {
//		panic(err)
//	}
//	var wg sync.WaitGroup
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		reader := bufio.NewReader(stdout)
//		for {
//			readString, err := reader.ReadString('\n')
//			if err != nil || err == io.EOF {
//				return
//			}
//			fmt.Print(readString)
//		}
//	}()
//	err = c.Start()
//	wg.Wait()
//	if err != nil {
//		panic(err)
//	}
//}

//type Charset string
//
//const (
//	UTF8    = Charset("UTF-8")
//	GB18030 = Charset("GB18030")
//)

//func main1() {
//	ctx, cancel := context.WithCancel(context.Background())
//	go func(cancelFunc context.CancelFunc) {
//		time.Sleep(5 * time.Second)
//		cancelFunc()
//	}(cancel)
//	err := Command(ctx, "ping www.baidu.com -n 6")
//	if err != nil {
//		panic(err)
//	}
//}
//
//func read(ctx context.Context, wg *sync.WaitGroup, std io.ReadCloser) {
//	reader := bufio.NewReader(std)
//	defer wg.Done()
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		default:
//			readString, err := reader.ReadString('\n')
//			if err != nil || err == io.EOF {
//				return
//			}
//			byte2String := ConvertByte2String([]byte(readString), "GBK")
//			fmt.Print(byte2String)
//		}
//	}
//}
//
////func ConvertByte2String(byte []byte, charset Charset) string {
////	var str string
////	switch charset {
////	case GB18030:
////		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
////		str = string(decodeBytes)
////	case UTF8:
////		fallthrough
////	default:
////		str = string(byte)
////	}
////	return str
////}
//
//func Command(ctx context.Context, cmd string) error {
//	//c := exec.CommandContext(ctx, "cmd", "/C", cmd) // windows
//	c := exec.CommandContext(ctx, "bash", "-c", cmd) // mac linux
//	stdout, err := c.StdoutPipe()
//	if err != nil {
//		return err
//	}
//	stderr, err := c.StderrPipe()
//	if err != nil {
//		return err
//	}
//	var wg sync.WaitGroup
//	// 因为有2个任务, 一个需要读取stderr 另一个需要读取stdout
//	wg.Add(2)
//	go read(ctx, &wg, stderr)
//	go read(ctx, &wg, stdout)
//	// 这里一定要用start,而不是run 详情请看下面的图
//	err = c.Start()
//	// 等待任务结束
//	wg.Wait()
//	return err
//}