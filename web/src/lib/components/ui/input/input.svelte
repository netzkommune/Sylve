<script lang="ts">
	import { cn } from '$lib/utils.js';
	import type { HTMLInputAttributes } from 'svelte/elements';
	import type { InputEvents } from './index.js';

	type $$Props = HTMLInputAttributes;
	type $$Events = InputEvents;

	let className: $$Props['class'] = undefined;
	export let value: $$Props['value'] = undefined;
	export { className as class };
	export let readonly: $$Props['readonly'] = undefined;

	// Add these for password visibility control
	export let type: $$Props['type'] = 'text';
	let showPassword = false;

	function handleFocus() {
		if (type === 'password') {
			showPassword = true;
		}
	}

	function handleBlur() {
		if (type === 'password') {
			showPassword = false;
		}
	}
</script>

<input
	class={cn(
		'border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-full rounded-md border px-3 py-2 text-sm file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
	bind:value
	{readonly}
	type={type === 'password' && showPassword ? 'text' : type}
	on:focus={handleFocus}
	on:blur={handleBlur}
	on:change
	on:click
	on:focusin
	on:focusout
	on:keydown
	on:keypress
	on:keyup
	on:mouseover
	on:mouseenter
	on:mouseleave
	on:mousemove
	on:paste
	on:input
	on:wheel|passive
	{...$$restProps}
/>
