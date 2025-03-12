import en from '$lib/locale/en.json';
import mal from '$lib/locale/mal.json';
import { _, addMessages, init } from 'svelte-i18n';
import { get } from 'svelte/store';

addMessages('en', en);
addMessages('mal', mal);

let savedLang: string = 'en';

if (typeof window !== 'undefined') {
	const stored = window.localStorage.getItem('language');
	try {
		const parsed = JSON.parse(stored || '');
		savedLang = parsed.value;
	} catch (e) {
		savedLang = 'en';
	}
}

init({
	initialLocale: savedLang,
	fallbackLocale: 'en'
});

export function getTranslation(key: string, fallback: string) {
	const translation = get(_)(key);
	return translation !== key ? translation : fallback;
}

export function getValidationError(s: string, section: string) {
	const split = s.split(' ');
	const keyF = split[0];
	const msgF = split.slice(1).join(' ');
	const faulted = getTranslation(`${section}.${keyF}`, keyF);
	let message = getTranslation(`validation_errors.${msgF}`, msgF);

	if (message === msgF) {
		const rawValidationErrors = get(_)('validation_errors');
		const validationErrors =
			typeof rawValidationErrors === 'object' && rawValidationErrors !== null
				? (rawValidationErrors as Record<string, string>)
				: {};

		message = validationErrors[msgF] || msgF;
	}

	return `${faulted} ${message}`;
}
