import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import AuthCard from '../components/AuthCard';

const SignUp = () => {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    fullName: '',
    email: '',
    password: '',
    companyName: '',
    companyWebsite: '',
    role: '',
    mrrRange: '',
    heardFrom: ''
  });

  const [errors, setErrors] = useState({});

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }));
    }
  };

  const validateForm = () => {
    const newErrors = {};

    // Required fields
    if (!formData.fullName.trim()) {
      newErrors.fullName = 'Full name is required';
    }

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    }

    if (!formData.companyName.trim()) {
      newErrors.companyName = 'Company name is required';
    }

    if (!formData.companyWebsite.trim()) {
      newErrors.companyWebsite = 'Company website is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    
    if (validateForm()) {
      console.log('Sign up data:', formData);
      // TODO: API call here
      
      // Redirect to onboarding
      navigate('/onboarding');
    }
  };

  const handleSignInClick = () => {
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900 flex flex-col items-center justify-center px-4 py-12">
      {/* Logo at top */}
      <div className="w-full max-w-2xl animate-fadeIn">
        <div className="flex justify-center mb-8">
          <img 
            src="/revalyze-logo.png" 
            alt="Revalyze" 
            className="h-14 w-auto"
          />
        </div>

        <AuthCard 
          title="Create your Revalyze account"
          subtitle="Start using AI to design your SaaS pricing"
        >
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Full Name */}
              <div>
                <label 
                  htmlFor="fullName" 
                  className="block text-sm font-semibold text-slate-300 mb-2"
                >
                  Full Name
                </label>
                <input
                  id="fullName"
                  name="fullName"
                  type="text"
                  value={formData.fullName}
                  onChange={handleChange}
                  className={`w-full px-4 py-3 rounded-xl bg-slate-900/50 border ${
                    errors.fullName ? 'border-red-500' : 'border-slate-700'
                  } text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all`}
                  placeholder="John Doe"
                />
                {errors.fullName && (
                  <p className="mt-1 text-sm text-red-400">{errors.fullName}</p>
                )}
              </div>

              {/* Email */}
              <div>
                <label 
                  htmlFor="email" 
                  className="block text-sm font-semibold text-slate-300 mb-2"
                >
                  Work Email
                </label>
                <input
                  id="email"
                  name="email"
                  type="email"
                  value={formData.email}
                  onChange={handleChange}
                  className={`w-full px-4 py-3 rounded-xl bg-slate-900/50 border ${
                    errors.email ? 'border-red-500' : 'border-slate-700'
                  } text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all`}
                  placeholder="you@company.com"
                />
                {errors.email && (
                  <p className="mt-1 text-sm text-red-400">{errors.email}</p>
                )}
              </div>

              {/* Password */}
              <div>
                <label 
                  htmlFor="password" 
                  className="block text-sm font-semibold text-slate-300 mb-2"
                >
                  Password
                </label>
                <input
                  id="password"
                  name="password"
                  type="password"
                  value={formData.password}
                  onChange={handleChange}
                  className={`w-full px-4 py-3 rounded-xl bg-slate-900/50 border ${
                    errors.password ? 'border-red-500' : 'border-slate-700'
                  } text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all`}
                  placeholder="Minimum 8 characters"
                />
                {errors.password && (
                  <p className="mt-1 text-sm text-red-400">{errors.password}</p>
                )}
              </div>

              {/* Two columns for company info */}
              <div className="grid md:grid-cols-2 gap-6">
                {/* Company Name */}
                <div>
                  <label 
                    htmlFor="companyName" 
                    className="block text-sm font-semibold text-slate-300 mb-2"
                  >
                    Company Name
                  </label>
                  <input
                    id="companyName"
                    name="companyName"
                    type="text"
                    value={formData.companyName}
                    onChange={handleChange}
                    className={`w-full px-4 py-3 rounded-xl bg-slate-900/50 border ${
                      errors.companyName ? 'border-red-500' : 'border-slate-700'
                    } text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all`}
                    placeholder="Acme Inc."
                  />
                  {errors.companyName && (
                    <p className="mt-1 text-sm text-red-400">{errors.companyName}</p>
                  )}
                </div>

                {/* Company Website */}
                <div>
                  <label 
                    htmlFor="companyWebsite" 
                    className="block text-sm font-semibold text-slate-300 mb-2"
                  >
                    Company Website
                  </label>
                  <input
                    id="companyWebsite"
                    name="companyWebsite"
                    type="url"
                    value={formData.companyWebsite}
                    onChange={handleChange}
                    className={`w-full px-4 py-3 rounded-xl bg-slate-900/50 border ${
                      errors.companyWebsite ? 'border-red-500' : 'border-slate-700'
                    } text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all`}
                    placeholder="acme.com"
                  />
                  {errors.companyWebsite && (
                    <p className="mt-1 text-sm text-red-400">{errors.companyWebsite}</p>
                  )}
                </div>
              </div>

              {/* Optional fields section */}
              <div className="pt-4 border-t border-slate-700">
                <p className="text-sm text-slate-400 mb-6">
                  Optional information (helps us personalize your experience)
                </p>

                <div className="space-y-6">
                  {/* Role */}
                  <div>
                    <label 
                      htmlFor="role" 
                      className="block text-sm font-semibold text-slate-300 mb-2"
                    >
                      Your Role <span className="text-slate-500 font-normal">(optional)</span>
                    </label>
                    <select
                      id="role"
                      name="role"
                      value={formData.role}
                      onChange={handleChange}
                      className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
                    >
                      <option value="">Select your role</option>
                      <option value="founder">Founder</option>
                      <option value="cofounder">Co-founder</option>
                      <option value="product">Product</option>
                      <option value="growth">Growth</option>
                      <option value="revenue">Revenue</option>
                      <option value="other">Other</option>
                    </select>
                  </div>

                  {/* Two columns for remaining optional */}
                  <div className="grid md:grid-cols-2 gap-6">
                    {/* MRR Range */}
                    <div>
                      <label 
                        htmlFor="mrrRange" 
                        className="block text-sm font-semibold text-slate-300 mb-2"
                      >
                        Current MRR <span className="text-slate-500 font-normal">(optional)</span>
                      </label>
                      <select
                        id="mrrRange"
                        name="mrrRange"
                        value={formData.mrrRange}
                        onChange={handleChange}
                        className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
                      >
                        <option value="">Select range</option>
                        <option value="<1k">&lt;$1k</option>
                        <option value="1k-10k">$1k–$10k</option>
                        <option value="10k-50k">$10k–$50k</option>
                        <option value="50k-100k">$50k–$100k</option>
                        <option value="100k+">$100k+</option>
                      </select>
                    </div>

                    {/* How did you hear */}
                    <div>
                      <label 
                        htmlFor="heardFrom" 
                        className="block text-sm font-semibold text-slate-300 mb-2"
                      >
                        How did you hear about us? <span className="text-slate-500 font-normal">(optional)</span>
                      </label>
                      <select
                        id="heardFrom"
                        name="heardFrom"
                        value={formData.heardFrom}
                        onChange={handleChange}
                        className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
                      >
                        <option value="">Select source</option>
                        <option value="linkedin">LinkedIn</option>
                        <option value="twitter">X (Twitter)</option>
                        <option value="friend">Friend</option>
                        <option value="search">Search</option>
                        <option value="other">Other</option>
                      </select>
                    </div>
                  </div>
                </div>
              </div>

              {/* Submit Button */}
              <button
                type="submit"
                className="w-full bg-gradient-to-r from-blue-500 to-indigo-600 text-white font-semibold py-3.5 rounded-xl hover:from-blue-600 hover:to-indigo-700 hover:scale-105 hover:shadow-xl hover:shadow-blue-500/30 transition-all duration-200"
              >
                Create Account
              </button>

              {/* Sign In Link */}
              <div className="text-center pt-4">
                <p className="text-sm text-slate-400">
                  Already have an account?{' '}
                  <button 
                    type="button"
                    onClick={handleSignInClick}
                    className="font-semibold text-blue-400 hover:text-blue-300 transition-colors"
                  >
                    Sign in
                  </button>
                </p>
              </div>
          </form>
        </AuthCard>
      </div>

      {/* Footer */}
      <div className="mt-8 text-center">
        <p className="text-slate-400 text-sm">
          © {new Date().getFullYear()} Revalyze B.V.
        </p>
      </div>
    </div>
  );
};

export default SignUp;

