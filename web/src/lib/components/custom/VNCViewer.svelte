<script lang="ts">
	import NoVNC from 'svelte-vnc';

	interface Props {
		vncPort: number;
		vncPassword: string;
	}

	let { vncPort, vncPassword }: Props = $props();

	let host = window.location.hostname;
	let port = window.location.port ? window.location.port : null;
	let path = `api/vnc/${encodeURIComponent(String(vncPort))}`;
</script>

<div class="h-full w-full">
	<NoVNC
		class="h-full w-full"
		{host}
		{port}
		{path}
		password={vncPassword}
		clearLocalStorage={true}
	/>
</div>
