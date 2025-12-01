import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';

const Navbar = () => {
  const [isScrolled, setIsScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 10);
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const scrollToSection = (sectionId) => {
    const element = document.getElementById(sectionId);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
  };

  return (
    <motion.nav
      initial={{ y: -100 }}
      animate={{ y: 0 }}
      transition={{ duration: 0.5 }}
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        isScrolled
          ? 'bg-slate-900/90 backdrop-blur-xl border-b border-slate-700'
          : 'bg-transparent'
      }`}
    >
      <div className="max-w-7xl mx-auto px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex-shrink-0">
            <button
              onClick={() => scrollToSection('hero')}
              className="flex items-center"
            >
              <img 
                src="/revalyze-logo.png" 
                alt="Revalyze" 
                className="h-12 w-auto"
              />
            </button>
          </div>

          <div className="hidden md:flex items-center space-x-1">
            <button
              onClick={() => scrollToSection('product')}
              className="px-4 py-2 text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition-all"
            >
              Product
            </button>
            <button
              onClick={() => scrollToSection('how-it-works')}
              className="px-4 py-2 text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition-all"
            >
              How it Works
            </button>
            <button
              onClick={() => scrollToSection('pricing')}
              className="px-4 py-2 text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition-all"
            >
              Pricing
            </button>
            
            <div className="flex items-center space-x-2 ml-6">
              <a
                href="http://localhost:3000"
                className="px-4 py-2 text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition-colors"
              >
                Login
              </a>
              <a
                href="http://localhost:3000/signup"
                className="px-5 py-2 text-sm font-medium bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl hover:shadow-lg hover:shadow-blue-500/50 transition-all hover:scale-105"
              >
                Sign Up
              </a>
            </div>
          </div>

          <div className="md:hidden flex items-center space-x-2">
            <a
              href="http://localhost:3000"
              className="px-3 py-1.5 text-sm font-medium text-slate-300"
            >
              Login
            </a>
            <a
              href="http://localhost:3000/signup"
              className="px-3 py-1.5 text-sm font-medium bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-lg"
            >
              Sign Up
            </a>
          </div>
        </div>
      </div>
    </motion.nav>
  );
};

export default Navbar;
