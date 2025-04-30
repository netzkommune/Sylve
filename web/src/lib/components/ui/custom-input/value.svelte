<script lang="ts">
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import { generateNanoId } from '$lib/utils/string';
	import type { FullAutoFill } from 'svelte/elements';

	interface Props {
		label: string;
		value: string | number;
		placeholder: string;
		autocomplete?: FullAutoFill | null | undefined;
		classes: string;
		type?: string;
	}

	let {
		value = $bindable(''),
		label = '',
		placeholder = '',
		autocomplete = 'off',
		classes = 'space-y-1',
		type = 'text'
	}: Props = $props();

	let nanoId = $state(generateNanoId(label));
</script>

<div class={classes}>
	<Label for={nanoId}>{label}</Label>
	<Input {type} id={nanoId} {placeholder} {autocomplete} bind:value />
</div>
