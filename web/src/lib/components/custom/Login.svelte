<script lang="ts">
	import { page } from '$app/stores';
	import { revokeJWT } from '$lib/api/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { hostname, language as langStore } from '$lib/stores/basic';
	import { getTranslation } from '$lib/utils/i18n';
	import { mode } from 'mode-watcher';
	import { _ } from 'svelte-i18n';

	interface Props {
		onLogin: (
			username: string,
			password: string,
			type: string,
			language: string,
			remember: boolean
		) => void;
	}

	let { onLogin }: Props = $props();

	let username = $state('');
	let password = $state('');
	let authType = $state('sylve');
	let language = $state('en');
	let remember = $state(false);

	$effect(() => {
		if ($page.url.search.includes('loggedOut')) {
			revokeJWT();
		}
	});
</script>

<div class="fixed inset-0 flex items-center justify-center bg-black bg-opacity-70 px-3">
	<Card.Root class="w-full max-w-lg rounded-lg shadow-lg">
		<Card.Header class="flex flex-row items-center justify-center gap-2">
			{#if $mode === 'dark'}
				<img src="/logo/white.svg" alt="Sylve Logo" class="mt-2 h-8 w-auto" />
			{:else}
				<img src="/logo/black.svg" alt="Sylve Logo" class="h-8 w-auto" />
			{/if}
			<p class="ml-2 text-xl font-medium tracking-[.45em] text-gray-800 dark:text-white">SYLVE</p>
		</Card.Header>

		<Card.Content class="space-y-4 p-6">
			<div class="flex items-center gap-2">
				<Label for="username" class="w-44">{$_('auth.username')}</Label>
				<Input
					id="username"
					class="h-8 w-full"
					type="text"
					placeholder={$_('common.example')}
					bind:value={username}
					required
				/>
			</div>
			<div class="flex items-center gap-2">
				<Label for="password" class="w-44">{$_('auth.password')}</Label>
				<Input
					id="password"
					type="password"
					placeholder="●●●●●●●●"
					class="h-8 w-full"
					bind:value={password}
					required
				/>
			</div>

			<div class="flex items-center gap-2">
				<Label for="realm" class="w-44">{$_('auth.realm')}</Label>
				<Select.Root
					onSelectedChange={(v) => {
						authType = v?.value as string;
					}}
					selected={{
						label: getTranslation(`auth.${authType}`, authType),
						value: authType
					}}
				>
					<Select.Trigger class="h-8 w-full">
						<Select.Value placeholder={$_('auth.select_authentication')} />
					</Select.Trigger>
					<Select.Content>
						<Select.Item value="pam">{$_('auth.pam')}</Select.Item>
						<Select.Item value="sylve">{$_('auth.sylve')}</Select.Item>
					</Select.Content>
				</Select.Root>
			</div>

			<div class="flex items-center gap-2">
				<Label for="language" class="w-44">{$_('auth.language')}</Label>
				<Select.Root
					onSelectedChange={(v) => (language = v?.value as string)}
					selected={{
						label: getTranslation(`common.languages.${language}`, language),
						value: language
					}}
				>
					<Select.Trigger class="h-8 w-full">
						<Select.Value placeholder={$_('auth.select_language')} />
					</Select.Trigger>
					<Select.Content>
						<Select.Item value="en">English</Select.Item>
						<Select.Item value="mal">Malayalam</Select.Item>
					</Select.Content>
				</Select.Root>
			</div>
		</Card.Content>

		<Card.Footer class="flex items-center justify-between px-6 py-4">
			<div class="flex items-center space-x-2">
				<Checkbox id="remember" bind:checked={remember} />
				<Label for="remember" class="text-sm font-medium">Remember Me</Label>
			</div>
			<Button
				onclick={() => {
					onLogin(username, password, authType, language, remember);
				}}
				size="sm"
				class="w-20 rounded-md bg-blue-700 text-white hover:bg-blue-600"
			>
				Login
			</Button>
		</Card.Footer>
	</Card.Root>
</div>
