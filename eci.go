package main

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/eci"
)


// 配置信息
var client *eci.Client

// 初始化配置信息
func InitEci() {
	//初始化client
	var err error
	client, err = eci.NewClientWithAccessKey(Config.Aliyun.RegionId, Config.Aliyun.AccessKey, Config.Aliyun.SecretKey)
	if err != nil {
		panic(err)
	}
}

// 创建容器组
func CreateContainerGroup () (containerGroupId string, err error)  {
	createContainerRequest := eci.CreateCreateContainerGroupRequest()

	// 新建网络和容器组
	createContainerRequest.RegionId = Config.Aliyun.RegionId
	createContainerRequest.ZoneId = Config.Aliyun.ZoneId
	createContainerRequest.SecurityGroupId = Config.Aliyun.SecurityGroupId
	createContainerRequest.VSwitchId = Config.Aliyun.VSwitchId

	createContainerRequest.ContainerGroupName = "minecraft"
	createContainerRequest.RestartPolicy = "Always"

	// 新建存储
	createContainerRequestVolume := make([]eci.CreateContainerGroupVolume, 1)
	volume1 := &eci.CreateContainerGroupNFSVolume{
		Path:"/",
		Server: Config.Container.Nfs,
	}
	createContainerRequestVolume[0].Name = "volume-ext"
	createContainerRequestVolume[0].Type ="NFSVolume"
	createContainerRequestVolume[0].NFSVolume=*volume1
	createContainerRequest.Volume = &createContainerRequestVolume

	// 新建容器组
	createContainerRequestContainer := make([]eci.CreateContainerGroupContainer, 1)
	createContainerRequestContainer[0].Image = Config.Container.Image
	createContainerRequestContainer[0].Name = "minecraft"
	// 选项
	createContainerRequestContainer[0].Cpu = requests.NewFloat(1)
	createContainerRequestContainer[0].Memory = requests.NewFloat(4)
	createContainerRequestContainer[0].ImagePullPolicy = "Always"
	// 挂载NFS
	createContainerGroupVolumeMount := make([]eci.CreateContainerGroupVolumeMount, 1)
	createContainerGroupVolumeMount[0] = eci.CreateContainerGroupVolumeMount{
		MountPath: "/data",
		ReadOnly: requests.NewBoolean(false),
		Name: "volume-ext",
	}
	createContainerRequestContainer[0].VolumeMount = &createContainerGroupVolumeMount
	// 暴露端口
	createContainerGroupPort := make([]eci.CreateContainerGroupPort, 1)
	createContainerGroupPort[0] = eci.CreateContainerGroupPort{
		Protocol: "TCP",
		Port: "25565",
	}
	createContainerRequestContainer[0].Port = &createContainerGroupPort

	createContainerRequest.Container = &createContainerRequestContainer

	//sdk-core默认的重试次数为3，在没有加幂等的条件下，资源创建的接口底层不需要自动重试
	client.GetConfig().MaxRetryTime = 0
	createContainerGroupResponse, err := client.CreateContainerGroup(createContainerRequest)
	if err != nil {
		return "", err
	}

	containerGroupId = createContainerGroupResponse.ContainerGroupId
	return containerGroupId, nil
}


// 删除容器组
func DeleteContainerGroup(containerGroupId string) {
	deleteContainerGroupRequest := eci.CreateDeleteContainerGroupRequest()
	deleteContainerGroupRequest.RegionId = Config.Aliyun.RegionId
	deleteContainerGroupRequest.ContainerGroupId = containerGroupId
	_, err := client.DeleteContainerGroup(deleteContainerGroupRequest)
	if err != nil {
		panic(err)
	}
}


// 查询容器组
func DescribeContainerGroup() (containerGroup *eci.DescribeContainerGroupsContainerGroup0, err error)  {
	describeContainerGroupsRequest := eci.CreateDescribeContainerGroupsRequest()

	describeContainerGroupsRequest.RegionId = Config.Aliyun.RegionId
	describeContainerGroupsRequest.ContainerGroupName = "minecraft"

	describeContainerGroupsResponse, err := client.DescribeContainerGroups(describeContainerGroupsRequest)
	if err != nil {
		return nil, err
	}

	describeContainerGroupNumber := len(describeContainerGroupsResponse.ContainerGroups)

	if describeContainerGroupNumber > 1 {
		for i := 1; i < describeContainerGroupNumber; i++ {
			// 删除重复容器，仅仅保留一个
			DeleteContainerGroup(describeContainerGroupsResponse.ContainerGroups[i].ContainerGroupId)
		}
	}else if describeContainerGroupNumber == 0 {
		return nil, nil
	}

	return &describeContainerGroupsResponse.ContainerGroups[0], nil
}