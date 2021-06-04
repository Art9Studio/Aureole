const axiosApiInstance = axios.create();

axiosApiInstance.interceptors.response.use((response) => {
    return response
}, async function (error) {
    const originalRequest = error.config;

    if (error.response.status === 401 && !originalRequest._retry) {
        originalRequest._retry = true;
        const refreshToken = localStorage.getItem('refresh');

        return axiosApiInstance.post('http://localhost:3000/emptyapi/refresh', {
                "refresh": refreshToken
            })
            .then(res => {
                localStorage.setItem('access', res.data.access);
                axiosApiInstance.defaults.headers.common['Authorization'] = 'Bearer ' + res.data.access;
                originalRequest.headers.Authorization = 'Bearer ' + res.data.access;

                return axios(originalRequest);
            })
            .catch(error => console.log(error.message));
    }
    return Promise.reject(error);
});

function getJsonFormData(form) {
    let formData = new FormData(form);
    let object = {};
    formData.forEach((value, key) => {object[key] = value});
    return JSON.stringify(object);
}

function sendForm(url, jsonFormData, successCallback = response => console.log(response)) {
    axiosApiInstance.post(url, jsonFormData, {
        baseURL: 'http://localhost:3000/emptyapi',
        headers: {'Content-Type': 'application/json'},
    })
        .then(successCallback)
        .catch(error => console.error(error.message));
}

function getIndex(url) {
    axiosApiInstance.get(url, {
        baseURL: 'http://localhost:8000/api',
    }).then(response => console.log(response));
}

document.getElementById('register-form').addEventListener('submit', (e) => {
    e.preventDefault();
    sendForm('/register', getJsonFormData(e.target))
});

document.getElementById('login-form').addEventListener('submit', (e) => {
    e.preventDefault();
    sendForm('/login', getJsonFormData(e.target), (response) => {
        console.log(response);
        localStorage.setItem('access', response.data.access);
        localStorage.setItem('refresh', response.data.refresh);
        axiosApiInstance.defaults.headers.common['Authorization'] = 'Bearer ' + response.data.access;
    })
});

document.getElementById('request').addEventListener('click', (e) => {
    e.preventDefault();
    getIndex('index/');
});

