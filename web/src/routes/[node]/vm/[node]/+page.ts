export async function load({ params }) {
    const childNode = params.childNode;

    return {
        data: {
            childNode
        }
    }
}