
    import defaultData, {key, pluralsRule} from 'virtual:wuchale/cn_simplified:main'
    const data = $state(defaultData)
    
    if (import.meta.hot) {
        import.meta.hot.on('virtual:wuchale/cn_simplified:main', newData => {
            for (let i = 0; i < newData.length; i++) {
                if (JSON.stringify(data[i]) !== JSON.stringify(newData[i])) {
                    data[i] = newData[i]
                }
            }
        })
        import.meta.hot.send('virtual:wuchale/cn_simplified:main')
    }

    export {key, pluralsRule}
    export default data
