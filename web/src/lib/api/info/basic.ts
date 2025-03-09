import { BasicInfoSchema, type BasicInfo } from '$lib/types/info/basic';
import { apiRequest } from '$lib/utils/http';

export async function getBasicInfo(): Promise<BasicInfo> {
	const data = await apiRequest('/info/basic', BasicInfoSchema, 'GET');
	return BasicInfoSchema.parse(data);
}
