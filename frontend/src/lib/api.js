
import { env } from '$env/dynamic/public';

console.log("API URL: ", env.PUBLIC_API_URL);
export const API = env.PUBLIC_API_URL || "http://localhost:8080/api";