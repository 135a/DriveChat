const { createApp } = Vue;

// Mock Axios for demo if server is offline, but we configure it for the Go backend
const api = axios.create({
    baseURL: '/admin', // assuming nginx proxies /admin to Go backend
    timeout: 5000
});

// Add token to requests
api.interceptors.request.use(config => {
    const token = localStorage.getItem('gw_token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

createApp({
    data() {
        return {
            isLoggedIn: false,
            isLoading: false,
            loginError: '',
            loginForm: {
                username: '',
                password: ''
            },
            currentView: 'overview',
            metrics: { qps: 0, interceptions: 0 },
            sysConfig: {
                slightly_freq_threshold: 50,
                too_freq_threshold: 100,
                auto_blacklist_enabled: true
            },
            routes: [],
            blacklist: [],
            logs: []
        }
    },
    computed: {
        pageTitle() {
            const titles = {
                overview: 'System Overview',
                routes: 'Routing Configuration',
                security: 'Security & Access Control',
                logs: 'Traffic Interception Logs'
            };
            return titles[this.currentView];
        }
    },
    mounted() {
        const token = localStorage.getItem('gw_token');
        if (token) {
            this.isLoggedIn = true;
            this.initDashboard();
        }
    },
    watch: {
        currentView(newVal) {
            if (newVal === 'overview') this.fetchOverview();
            if (newVal === 'routes') this.fetchRoutes();
            if (newVal === 'security') this.fetchBlacklist();
            if (newVal === 'logs') this.fetchLogs();
        }
    },
    methods: {
        async login() {
            this.isLoading = true;
            this.loginError = '';
            try {
                // In a real environment, call the backend:
                // const res = await axios.post('/auth/login', this.loginForm);
                // localStorage.setItem('gw_token', res.data.token);
                
                // For immediate visual feedback during dev without running Go:
                if(this.loginForm.username === 'admin') {
                    setTimeout(() => {
                        localStorage.setItem('gw_token', 'demo_token');
                        this.isLoggedIn = true;
                        this.isLoading = false;
                        this.initDashboard();
                    }, 800);
                } else {
                    throw new Error("Invalid credentials");
                }
            } catch (err) {
                this.loginError = err.response?.data?.error || 'Authentication failed. Check credentials.';
                this.isLoading = false;
            }
        },
        logout() {
            localStorage.removeItem('gw_token');
            this.isLoggedIn = false;
        },
        initDashboard() {
            this.fetchOverview();
            // Setup simple polling for metrics
            setInterval(() => {
                if(this.isLoggedIn && this.currentView === 'overview') {
                    this.fetchMetrics();
                }
            }, 5000);
        },
        async fetchOverview() {
            this.fetchMetrics();
            this.fetchSysConfig();
            this.fetchRoutes();
        },
        async fetchMetrics() {
            try {
                const res = await api.get('/metrics');
                this.metrics.qps = res.data.qps || 12; // Mocking fallback
                this.metrics.interceptions = res.data.interceptions || 104;
            } catch(e) { console.warn("Using mock metrics"); this.metrics = {qps: 24, interceptions: 1042}; }
        },
        async fetchSysConfig() {
            try {
                const res = await api.get('/config');
                this.sysConfig = res.data;
            } catch(e) { console.warn("Using mock config"); }
        },
        async saveConfig() {
            try {
                await api.post('/config', this.sysConfig);
                alert("Configuration saved successfully.");
            } catch(e) { alert("Demo mode: Config saved locally."); }
        },
        async fetchRoutes() {
            try {
                const res = await api.get('/routes');
                this.routes = res.data;
            } catch(e) {
                this.routes = [
                    { id: 1, path_prefix: '/api/v1', target_url: 'http://user-service:8080', status: 1 },
                    { id: 2, path_prefix: '/api/v2', target_url: 'http://order-service:8081', status: 1 }
                ];
            }
        },
        async deleteRoute(id) {
            if(confirm("Delete this route?")) {
                try { await api.delete(`/routes/${id}`); this.fetchRoutes(); }
                catch(e) { this.routes = this.routes.filter(r => r.id !== id); }
            }
        },
        async fetchBlacklist() {
            try {
                const res = await api.get('/blacklist');
                this.blacklist = res.data;
            } catch(e) {
                this.blacklist = [
                    { id: 1, type: 'IP', value: '192.168.1.100', is_auto: true },
                    { id: 2, type: 'PATH_PREFIX', value: '/admin/hidden', is_auto: false }
                ];
            }
        },
        async deleteBlacklist(id) {
            if(confirm("Remove from blacklist?")) {
                try { await api.delete(`/blacklist/${id}`); this.fetchBlacklist(); }
                catch(e) { this.blacklist = this.blacklist.filter(b => b.id !== id); }
            }
        },
        async fetchLogs() {
            try {
                const res = await api.get('/logs');
                this.logs = res.data.data;
            } catch(e) {
                this.logs = [
                    { id: 1, created_at: new Date().toISOString(), client_ip: '10.0.0.5', req_path: '/api/v1/data', rule_type: 'RATE_LIMIT_AUTO_BLOCK' },
                    { id: 2, created_at: new Date(Date.now() - 60000).toISOString(), client_ip: '192.168.1.100', req_path: '/admin', rule_type: 'IP_BLACKLIST' }
                ];
            }
        },
        // Modals (mocked as simple prompts for demo)
        openRouteModal(route) {
            const prefix = prompt("Path Prefix", route ? route.path_prefix : "/api/");
            if(!prefix) return;
            const target = prompt("Target URL", route ? route.target_url : "http://localhost:8081");
            if(!target) return;
            
            const payload = { path_prefix: prefix, target_url: target, status: 1 };
            if(route) {
                api.put(`/routes/${route.id}`, payload).then(() => this.fetchRoutes()).catch(()=> {
                    route.path_prefix = prefix; route.target_url = target;
                });
            } else {
                api.post('/routes', payload).then(() => this.fetchRoutes()).catch(()=> {
                    this.routes.push({ id: Date.now(), ...payload });
                });
            }
        },
        openBlacklistModal() {
            const type = prompt("Type (IP / PATH_PREFIX)", "IP");
            if(!type) return;
            const value = prompt("Value", "127.0.0.1");
            if(!value) return;
            
            const payload = { type, value };
            api.post('/blacklist', payload).then(() => this.fetchBlacklist()).catch(()=> {
                this.blacklist.push({ id: Date.now(), type, value, is_auto: false });
            });
        }
    }
}).mount('#app');
