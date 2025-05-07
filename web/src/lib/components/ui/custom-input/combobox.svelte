<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Command from '$lib/components/ui/command/index.js';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { cn } from '$lib/utils.js';
	import Icon from '@iconify/svelte';

	interface Props {
		open: boolean;
		label: string;
		value: string;
		data: { value: string; label: string }[];
		onValueChange?: (value: string) => void;
		placeholder?: string;
		disabled?: boolean;
		classes?: string;
		width?: string;
	}

	let {
		open = $bindable(false),
		label = '',
		value = $bindable(''),
		data = [],
		onValueChange = () => {},
		placeholder = '',
		disabled = false,
		classes = 'space-y-1',
		width = 'w-1/2'
	}: Props = $props();

	let search = $state('');
	let lowerCaseLabel = $state(label.toLowerCase());
	let filteredData = $derived.by(() => {
		if (!search) return data;
		const q = search.toLowerCase();
		return data.filter(
			({ label, value }) => label.toLowerCase().includes(q) || value.toLowerCase().includes(q)
		);
	});

	function selectItem(val: string) {
		if (value === val) {
			value = '';
			onValueChange('');
		} else {
			value = val;
			onValueChange(val);
		}
		search = '';
		open = false;
	}

	let selectedLabel = $derived.by(() => {
		const selected = data.find((item) => item.value === value);
		return selected ? selected.label : '';
	});
</script>

<div class={classes}>
	<Label class="w-24 whitespace-nowrap text-sm" for={lowerCaseLabel}>{label}</Label>
	<Popover.Root bind:open>
		<Popover.Trigger asChild let:builder>
			<Button
				builders={[builder]}
				variant="outline"
				role="combobox"
				aria-expanded={open}
				class="w-full justify-between"
				{disabled}
			>
				{selectedLabel || placeholder}
				<Icon icon="lucide:chevrons-up-down" class="ml-2 h-4 w-4 shrink-0 opacity-50" />
			</Button>
		</Popover.Trigger>

		<Popover.Content class="{width} p-0">
			<Command.Root shouldFilter={false}>
				<Command.Input bind:value={search} placeholder={placeholder || 'Search...'} />
				<Command.Empty>No data</Command.Empty>
				<div class="max-h-64 overflow-y-auto">
					<Command.Group>
						{#each filteredData as element}
							<Command.Item
								value={element.value}
								onSelect={() => selectItem(element.value)}
								onkeydown={(e) => {
									if (e.key === 'Enter') selectItem(element.value);
								}}
							>
								<Icon
									icon="lucide:check"
									class={cn('mr-2 h-4 w-4', value !== element.value && 'text-transparent')}
								/>
								{element.label}
							</Command.Item>
						{/each}
					</Command.Group>
				</div>
			</Command.Root>
		</Popover.Content>
	</Popover.Root>
</div>
