import axios from 'axios';
import { useState } from 'react';

const Urlshortner = () => {
    const [inputdata, setinputdata] = useState('');
    const [shortenURl, setshortenURL] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleClick = async () => {
        if (!inputdata.trim()) {
            setError('Please enter a URL');
            return;
        }

        setLoading(true);
        setError('');

        try {
            const response = await axios.post(
                'http://localhost:8080/create',
                { long_url: inputdata },
                { headers: { 'Content-type': 'application/json' } }
            );
            setshortenURL(response.data.short_url || response.data);
        } catch (error) {
            console.error(error);
            setError('Failed to shorten URL. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex flex-col items-center gap-4 p-4 max-w-md mx-auto">
            <div className="flex flex-col w-full gap-4">
                <div className="flex w-full gap-2">
                    <input
                        onChange={(e) => setinputdata(e.target.value)}
                        value={inputdata}
                        type="url"
                        placeholder="Enter your URL"
                        className="flex-1 px-4 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                        style={{ minWidth: '150px' }}
                    />
                    <button
                        onClick={handleClick}
                        disabled={loading}
                        className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-blue-300"
                    >
                        {loading ? 'Shortening...' : 'Submit'}
                    </button>
                </div>

                {error && <p className="text-red-500 text-sm">{error}</p>}

                {shortenURl && (
                    <div className="w-full p-4 bg-gray-50 rounded border relative">
                        <button
                            onClick={() =>
                                navigator.clipboard.writeText(shortenURl)
                            }
                            className="absolute top-2 right-2 p-1 text-gray-500 hover:text-gray-700"
                            title="Copy to clipboard"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                className="h-5 w-5"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth={2}
                                    d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                                />
                            </svg>
                        </button>
                        <p className="text-sm text-gray-600 mb-2">
                            Shortened URL:
                        </p>
                        <div className="relative">
                            <a
                                href={
                                    shortenURl.startsWith('http')
                                        ? shortenURl
                                        : `https://${shortenURl}`
                                }
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-blue-500 hover:underline break-all pr-6"
                            >
                                {shortenURl}
                            </a>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default Urlshortner;
