import axios from "axios";

const client = axios.create({
  baseURL: process.env.VUE_APP_API_URL || "http://localhost:8000",
  timeout: 15000,
});

// Attach X-API-Key from localStorage on every request
client.interceptors.request.use((config) => {
  const key = localStorage.getItem("api_key");
  if (key) {
    config.headers["X-API-Key"] = key;
  }
  return config;
});

export const statsAPI = {
  get: () => client.get("/stats"),
};

export const clientsAPI = {
  list: () => client.get("/clients/"),
  get: (id) => client.get(`/clients/${id}`),
  register: (data) => client.post("/clients/register", data),
  delete: (id) => client.delete(`/clients/${id}`),
  heartbeat: (id, data) => client.post(`/clients/${id}/heartbeat`, data),
};

export const updatesAPI = {
  list: () => client.get("/updates/"),
  get: (id) => client.get(`/updates/${id}`),
  latest: () => client.get("/updates/latest"),
  create: (data) => client.post("/updates/", data),
  delete: (id) => client.delete(`/updates/${id}`),
  confirm: (id, data) => client.post(`/updates/${id}/confirm`, data),
};

export default client;
