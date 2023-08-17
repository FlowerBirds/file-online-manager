package util

import (
	"log"
	"os/exec"
)

func ExecuteCommand(command string, args ...string) error {
	log.Println(command, args)
	cmd := exec.Command(command, args...)

	// 创建管道用于获取命令输出
	//pipeReader, pipeWriter := io.Pipe()
	//cmd.Stderr = os.Stderr
	//cmd.Stdout = pipeWriter

	// 开启协程读取命令输出并打印
	//go func() {
	//	defer pipeReader.Close()
	//	scanner := bufio.NewScanner(pipeReader)
	//	for scanner.Scan() {
	//		fmt.Println(scanner.Text())
	//	}
	//}()

	// 执行命令
	err := cmd.Run()
	//pipeWriter.Close()
	//if err != nil {
	//	return fmt.Errorf("执行命令失败：%v", err)
	//}
	return err
}
