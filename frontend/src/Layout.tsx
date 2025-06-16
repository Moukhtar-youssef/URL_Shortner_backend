import { Link, Outlet } from 'react-router';

const Layout = () => {
    return (
        <>
            <nav className="sticky top-0 z-50 ">
                <ul>
                    <li>
                        <Link to="/">Home</Link>
                    </li>
                    <li>
                        <Link to="/about">about</Link>
                    </li>
                </ul>
            </nav>

            <Outlet />
        </>
    );
};

export default Layout;
