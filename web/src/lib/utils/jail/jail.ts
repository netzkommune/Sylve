import type { CreateData } from '$lib/types/jail/jail';
import { toast } from 'svelte-sonner';
import { isValidVMName } from '../string';

export function isValidCreateData(modal: CreateData): boolean {
	const toastConfig: Record<string, unknown> = {
		duration: 3000,
		position: 'bottom-center'
	};

	if (!isValidVMName(modal.name)) {
		toast.error('Invalid name', toastConfig);
		return false;
	}

	if (modal.id < 1 || modal.id > 9999) {
		toast.error('Invalid ID', toastConfig);
		return false;
	}

	if (modal.description && (modal.description.length < 1 || modal.description.length > 1024)) {
		toast.error('Invalid description', toastConfig);
		return false;
	}

	if (modal.storage.dataset.length < 1) {
		toast.error('No storage dataset selected', toastConfig);
		return false;
	}

	if (modal.storage.base.length < 1) {
		toast.error('No system base selected', toastConfig);
		return false;
	}

	return true;
}
