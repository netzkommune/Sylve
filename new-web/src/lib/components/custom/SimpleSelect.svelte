<script lang="ts">
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Select from '$lib/components/ui/select/index.js';

	interface Props {
		label?: string;
		placeholder?: string;
		options: Array<{ value: string; label: string }>;
		value: string;
		classes?: {
			parent?: string;
			label?: string;
		};
		onChange: (value: string) => void;
		disabled?: boolean;
	}

	let {
		label = 'Select',
		placeholder = 'Select an option',
		options,
		classes = { parent: 'flex-1 space-y-1', label: 'w-24 whitespace-nowrap text-sm' },
		value = $bindable(),
		onChange,
		disabled = false
	}: Props = $props();
</script>

<div class={classes.parent}>
	<Label class={classes.label}>{label}</Label>
	<Select.Root
		type="single"
		bind:value
		onValueChange={() => {
			onChange(value);
		}}
		{disabled}
	>
		<Select.Trigger class="w-full">
			{value ? options.find((o) => o.value === value)?.label : placeholder}
		</Select.Trigger>
		<Select.Content>
			{#each options as option (option.value)}
				<Select.Item value={option.value} label={option.label}>
					{option.label}
				</Select.Item>
			{/each}
		</Select.Content>
	</Select.Root>
</div>
