import { BasicInfoSchema, type BasicInfo } from '$lib/types/info/basic';
import { apiRequest } from '$lib/utils/http';

export async function getBasicInfo(): Promise<BasicInfo> {
	return await apiRequest('/info/basic', BasicInfoSchema, 'GET');
}
