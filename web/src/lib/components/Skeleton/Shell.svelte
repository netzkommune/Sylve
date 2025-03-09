<script lang="ts">
	import * as Resizable from '$lib/components/ui/resizable';
	import Header from '../custom/Header.svelte';

	import BottomPanel from '$lib/components/Skeleton/BottomPanel.svelte';
	import LeftPanel from '$lib/components/Skeleton/LeftPanel.svelte';
	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();
</script>

<div class="flex min-h-screen w-full flex-col">
	<Header />
	<main class="flex flex-1 flex-col">
		<div class="h-[95vh] w-full md:h-[96vh]">
			<Resizable.PaneGroup direction="vertical">
				<Resizable.Pane defaultSize={70}>
					<Resizable.PaneGroup direction="horizontal">
						<Resizable.Pane defaultSize={13} minSize={5}>
							<LeftPanel />
						</Resizable.Pane>

						<Resizable.Handle withHandle />

						<Resizable.Pane defaultSize={87}>
							{@render children?.()}
						</Resizable.Pane>
					</Resizable.PaneGroup>
				</Resizable.Pane>
				<Resizable.Handle withHandle />

				<Resizable.Pane defaultSize={30} class="h-full min-h-20">
					<BottomPanel />
				</Resizable.Pane>
			</Resizable.PaneGroup>
		</div>
	</main>
</div>
