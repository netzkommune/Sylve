import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { apiRequest } from '$lib/utils/http';

/* 
type CreateVMRequest struct {
	Name                 string   `json:"name" binding:"required"`
	VMID                 *int     `json:"vmId" binding:"required"`
	Description          string   `json:"description"`
	StorageType          string   `json:"storageType" binding:"required"`
	StorageDataset       string   `json:"storageDataset" binding:"required"`
	StorageSize          *int64   `json:"storageSize" binding:"required"`
	StorageEmulationType string   `json:"storageEmulationType"`
	NetworkSwitch        string   `json:"networkSwitch"`
	NetworkMAC           string   `json:"networkMAC"`
	CPUSockets           int      `json:"cpuSockets" binding:"required"`
    CPUCores             int      `json:"cpuCores" binding:"required"
	CPUThreads           int      `json:"cpuThreads" binding:"required"``
	RAM                  int      `json:"ram" binding:"required"`
	PCIDevices           []string `json:"pciDevices"`
	VNCPort              int      `json:"vncPort" binding:"required"`
	VNCPassword          string   `json:"vncPassword"`
	VNCResolution        string   `json:"vncResolution"`
	StartAtBoot          *bool    `json:"startAtBoot" binding:"required"`
	StartOrder           int      `json:"startOrder"`
}
*/
export async function newVM(
	name: string,
	vmId: number,
	storageType: string,
	storageDataset: string,
	storageSize: number,
	emulationType: string,
	switchId: number,
	macAddress: string,
	cpuSockets: number,
	cpuCores: number,
	cpuThreads: number,
	ram: number
): Promise<APIResponse> {
	return await apiRequest('/vm/create', APIResponseSchema, 'POST', {
		name,
		vmId,
		storageType,
		storageDataset,
		storageSize,
		emulationType,
		switchId,
		macAddress,
		cpuSockets,
		cpuCores,
		cpuThreads,
		ram
	});
}
