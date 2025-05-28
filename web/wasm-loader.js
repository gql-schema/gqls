// WASM loader and interface for the GQL to Neo4j converter
class WASMConverter {
    constructor() {
        this.wasmModule = null;
        this.loaded = false;
        this.loading = false;
    }

    async load() {
        if (this.loaded) return;
        if (this.loading) {
            // Wait for the current loading to complete
            while (this.loading && !this.loaded) {
                await new Promise(resolve => setTimeout(resolve, 100));
            }
            return;
        }

        this.loading = true;

        try {
            // Check if Go WASM support is available
            if (typeof Go === 'undefined') {
                throw new Error('Go WASM runtime not loaded. Make sure wasm_exec.js is included.');
            }

            // Load the Go WASM support script
            const go = new Go();

            // Load the WASM module
            const result = await WebAssembly.instantiateStreaming(
                fetch('./main.wasm'),
                go.importObject
            );

            // Run the Go program
            go.run(result.instance);

            this.loaded = true;
            this.loading = false;
            console.log('WASM module loaded successfully');
        } catch (error) {
            this.loading = false;
            console.error('Failed to load WASM module:', error);
            throw error;
        }
    }

    async convert(gql, neo4jVersion, apocEnabled) {
        if (!this.loaded) {
            await this.load();
        }

        if (typeof window.convertGQL !== 'function') {
            throw new Error('WASM module not properly initialized. The convertGQL function is not available.');
        }

        const request = JSON.stringify({
            gql,
            neo4jVersion: neo4jVersion || 'latest',
            apocEnabled: Boolean(apocEnabled)
        });

        try {
            return window.convertGQL(request);
        } catch (error) {
            console.error('WASM conversion error:', error);
            throw error;
        }
    }

    getVersion() {
        if (!this.loaded) {
            return 'Version not available (WASM not loaded)';
        }

        if (typeof window.getVersionInfo !== 'function') {
            return 'Version not available (function not found)';
        }

        try {
            return window.getVersionInfo();
        } catch (error) {
            console.error('WASM version error:', error);
            return 'Version not available (error)';
        }
    }
}

// Export for use in other scripts
window.WASMConverter = WASMConverter;
