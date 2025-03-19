<script lang="ts">
	import Header from '$lib/components/custom/Header.svelte';
	import * as Resizable from '$lib/components/ui/resizable';

	import Terminal from '$lib/components/custom/Terminal.svelte';
	import BottomPanel from '$lib/components/Skeleton/BottomPanel.svelte';
	import LeftPanel from '$lib/components/Skeleton/LeftPanel.svelte';
	import { paneSizes } from '$lib/stores/basic';
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
				<Resizable.Pane defaultSize={$paneSizes.main} onResize={(e) => ($paneSizes.main = e)}>
					<Resizable.PaneGroup direction="horizontal">
						<Resizable.Pane
							defaultSize={$paneSizes.left}
							minSize={5}
							onResize={(e) => ($paneSizes.left = e)}
						>
							<LeftPanel />
						</Resizable.Pane>

						<Resizable.Handle withHandle />

						<Resizable.Pane>
							{@render children?.()}
						</Resizable.Pane>
					</Resizable.PaneGroup>
				</Resizable.Pane>

				<Resizable.Handle withHandle />

				<Resizable.Pane
					defaultSize={$paneSizes.bottom}
					class="h-full min-h-20"
					onResize={(e) => ($paneSizes.bottom = e)}
				>
					<BottomPanel />
				</Resizable.Pane>
			</Resizable.PaneGroup>
		</div>

		<Terminal />
	</main>
</div>
