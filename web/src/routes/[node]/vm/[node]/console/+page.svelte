<script lang="ts">
	interface Data {
		port: number;
		password: string;
	}

	let { data }: { data: Data } = $props();
	let path = $derived(`/api/vnc/${encodeURIComponent(String(data.port))}`);
</script>

<div class="flex h-full w-full flex-col">
	<iframe class="w-full flex-1" src={`/vnc/vnc.html?path=${path}&password=${data.password}`}
	></iframe>
</div>
