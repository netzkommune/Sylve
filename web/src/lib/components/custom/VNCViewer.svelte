<script lang="ts">
	interface Props {
		vncPort: number;
		vncPassword: string;
	}

	let { vncPort, vncPassword }: Props = $props();

	const host = window.location.hostname;
	const port = window.location.port;
	const noVncBase = `https://${host}:${port}/novnc/vnc.html`;
	const noVncParams = [
		`host=${encodeURIComponent(host)}`,
		`port=${encodeURIComponent(port)}`,
		`path=api/vnc/${encodeURIComponent(String(vncPort))}`,
		`password=${encodeURIComponent(vncPassword)}`,
		`autoconnect=true`,
		`reconnect=true`
	].join('&');

	const iframeSrc = noVncBase + '?' + noVncParams;
</script>

<iframe
	title="noVNC session"
	src={iframeSrc}
	allowfullscreen
	class="h-full w-full rounded-lg border border-gray-300 shadow-lg"
/>
