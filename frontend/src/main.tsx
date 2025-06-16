import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import App from './pages/App.tsx';
import { BrowserRouter, Route, Routes } from 'react-router';
import About from './pages/About.tsx';
import Layout from './Layout.tsx';
import NoPage from './pages/NoPage.tsx';

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<Layout />}>
                    <Route index element={<App />} />
                    <Route path="about" element={<About />} />
                    <Route path="*" element={<NoPage />} />
                </Route>
            </Routes>
        </BrowserRouter>
    </StrictMode>
);
