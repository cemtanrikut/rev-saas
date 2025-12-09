import { motion } from 'framer-motion';

const SimulationShowcase = () => {
  return (
    <section className="py-20 md:py-28 bg-gradient-to-b from-slate-900 via-slate-900 to-slate-800 relative overflow-hidden">
      {/* Background decorative elements */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 -left-20 w-96 h-96 bg-emerald-500/10 rounded-full blur-3xl"></div>
        <div className="absolute bottom-1/4 -right-20 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl"></div>
      </div>

      <div className="max-w-7xl mx-auto px-6 lg:px-8 relative">
        {/* Section Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          viewport={{ once: true }}
          className="text-center mb-12 md:mb-16"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 bg-emerald-500/10 border border-emerald-500/30 rounded-full text-sm font-medium text-emerald-400 mb-6">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            Key Feature
          </div>
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-6 tracking-tight">
            What happens if you <br className="hidden md:block" />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-blue-400">
              change your prices?
            </span>
          </h2>
          <p className="text-xl md:text-2xl text-slate-400 max-w-3xl mx-auto leading-relaxed">
            Run pricing simulations before you ship. See the projected impact on revenue and customers — without the guesswork.
          </p>
        </motion.div>

        {/* Main Content Grid */}
        <div className="grid lg:grid-cols-12 gap-8 lg:gap-12 items-center">
          {/* Left side - Features */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            viewport={{ once: true }}
            className="lg:col-span-5 space-y-6"
          >
            <div className="flex items-start gap-4 p-5 bg-slate-800/60 rounded-2xl border border-slate-700/50 backdrop-blur-sm hover:border-emerald-500/30 transition-colors">
              <div className="w-12 h-12 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl flex items-center justify-center flex-shrink-0 shadow-lg shadow-emerald-500/25">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-white mb-1">3 Risk Scenarios</h3>
                <p className="text-slate-400">Conservative, Base, and Aggressive projections for every price change.</p>
              </div>
            </div>

            <div className="flex items-start gap-4 p-5 bg-slate-800/60 rounded-2xl border border-slate-700/50 backdrop-blur-sm hover:border-blue-500/30 transition-colors">
              <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl flex items-center justify-center flex-shrink-0 shadow-lg shadow-blue-500/25">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-white mb-1">MRR & ARR Projections</h3>
                <p className="text-slate-400">See exactly how your monthly and annual revenue could change.</p>
              </div>
            </div>

            <div className="flex items-start gap-4 p-5 bg-slate-800/60 rounded-2xl border border-slate-700/50 backdrop-blur-sm hover:border-purple-500/30 transition-colors">
              <div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl flex items-center justify-center flex-shrink-0 shadow-lg shadow-purple-500/25">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-white mb-1">Customer Impact</h3>
                <p className="text-slate-400">Understand potential customer loss or gain for each scenario.</p>
              </div>
            </div>

            <div className="flex items-start gap-4 p-5 bg-slate-800/60 rounded-2xl border border-slate-700/50 backdrop-blur-sm hover:border-amber-500/30 transition-colors">
              <div className="w-12 h-12 bg-gradient-to-br from-amber-500 to-amber-600 rounded-xl flex items-center justify-center flex-shrink-0 shadow-lg shadow-amber-500/25">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-white mb-1">AI Executive Summary</h3>
                <p className="text-slate-400">Get strategic insights written for decision-makers, not spreadsheets.</p>
              </div>
            </div>
          </motion.div>

          {/* Right side - Mock UI */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            whileInView={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
            viewport={{ once: true }}
            className="lg:col-span-7"
          >
            <div className="relative">
              {/* Glow effect behind the card */}
              <div className="absolute -inset-4 bg-gradient-to-r from-emerald-500/20 via-blue-500/20 to-purple-500/20 rounded-3xl blur-2xl opacity-60"></div>
              
              {/* Main Card */}
              <div className="relative bg-slate-800/90 backdrop-blur-xl rounded-3xl border border-slate-700 shadow-2xl overflow-hidden">
                {/* Card Header */}
                <div className="px-6 py-4 border-b border-slate-700 bg-slate-800/50">
                  <div className="flex items-center gap-3">
                    <div className="flex gap-1.5">
                      <div className="w-3 h-3 rounded-full bg-red-500"></div>
                      <div className="w-3 h-3 rounded-full bg-amber-500"></div>
                      <div className="w-3 h-3 rounded-full bg-emerald-500"></div>
                    </div>
                    <span className="text-sm text-slate-400">Pricing Simulation</span>
                  </div>
                </div>

                {/* Card Content */}
                <div className="p-6">
                  {/* Simulation Header */}
                  <div className="flex items-center justify-between mb-6 pb-6 border-b border-slate-700/50">
                    <div>
                      <div className="text-sm text-slate-400 mb-1">Pro Plan</div>
                      <div className="flex items-center gap-3">
                        <span className="text-2xl text-slate-400 line-through">$29</span>
                        <svg className="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                        </svg>
                        <span className="text-2xl text-white font-bold">$39</span>
                        <span className="text-slate-500">/month</span>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-2xl font-bold text-emerald-400">+34.5%</div>
                      <div className="inline-flex items-center gap-1.5 px-3 py-1 bg-amber-500/20 border border-amber-500/30 rounded-full">
                        <div className="w-2 h-2 bg-amber-400 rounded-full"></div>
                        <span className="text-sm text-amber-400 font-medium">Medium Risk</span>
                      </div>
                    </div>
                  </div>

                  {/* Scenario Cards */}
                  <div className="grid grid-cols-3 gap-4 mb-6">
                    {/* Conservative */}
                    <div className="bg-emerald-500/10 border border-emerald-500/30 rounded-xl p-4 hover:bg-emerald-500/15 transition-colors">
                      <div className="flex items-center gap-2 mb-3">
                        <div className="w-2 h-2 bg-emerald-400 rounded-full"></div>
                        <span className="text-sm font-semibold text-emerald-400">Conservative</span>
                      </div>
                      <div className="space-y-2">
                        <div>
                          <div className="text-xs text-slate-500 uppercase tracking-wide">Customer Loss</div>
                          <div className="text-sm text-white font-medium">2% – 4%</div>
                        </div>
                        <div>
                          <div className="text-xs text-slate-500 uppercase tracking-wide">Projected MRR</div>
                          <div className="text-lg text-white font-bold">$11.2K</div>
                        </div>
                        <div className="pt-2 border-t border-emerald-500/20">
                          <span className="text-xs px-2 py-0.5 bg-emerald-500/20 text-emerald-400 rounded-full">Low Risk</span>
                        </div>
                      </div>
                    </div>

                    {/* Base */}
                    <div className="bg-blue-500/10 border border-blue-500/30 rounded-xl p-4 hover:bg-blue-500/15 transition-colors">
                      <div className="flex items-center gap-2 mb-3">
                        <div className="w-2 h-2 bg-blue-400 rounded-full"></div>
                        <span className="text-sm font-semibold text-blue-400">Base</span>
                      </div>
                      <div className="space-y-2">
                        <div>
                          <div className="text-xs text-slate-500 uppercase tracking-wide">Customer Loss</div>
                          <div className="text-sm text-white font-medium">4% – 7%</div>
                        </div>
                        <div>
                          <div className="text-xs text-slate-500 uppercase tracking-wide">Projected MRR</div>
                          <div className="text-lg text-white font-bold">$10.8K</div>
                        </div>
                        <div className="pt-2 border-t border-blue-500/20">
                          <span className="text-xs px-2 py-0.5 bg-amber-500/20 text-amber-400 rounded-full">Medium Risk</span>
                        </div>
                      </div>
                    </div>

                    {/* Aggressive */}
                    <div className="bg-orange-500/10 border border-orange-500/30 rounded-xl p-4 hover:bg-orange-500/15 transition-colors">
                      <div className="flex items-center gap-2 mb-3">
                        <div className="w-2 h-2 bg-orange-400 rounded-full"></div>
                        <span className="text-sm font-semibold text-orange-400">Aggressive</span>
                      </div>
                      <div className="space-y-2">
                        <div>
                          <div className="text-xs text-slate-500 uppercase tracking-wide">Customer Loss</div>
                          <div className="text-sm text-white font-medium">7% – 12%</div>
                        </div>
                        <div>
                          <div className="text-xs text-slate-500 uppercase tracking-wide">Projected MRR</div>
                          <div className="text-lg text-white font-bold">$10.2K</div>
                        </div>
                        <div className="pt-2 border-t border-orange-500/20">
                          <span className="text-xs px-2 py-0.5 bg-red-500/20 text-red-400 rounded-full">High Risk</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* AI Insight */}
                  <div className="bg-gradient-to-br from-purple-500/10 via-blue-500/10 to-emerald-500/10 border border-purple-500/20 rounded-xl p-5">
                    <div className="flex items-center gap-2 mb-3">
                      <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-blue-500 rounded-lg flex items-center justify-center">
                        <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                        </svg>
                      </div>
                      <span className="text-sm font-semibold text-purple-400">AI Pricing Insight</span>
                    </div>
                    <p className="text-slate-300 text-sm leading-relaxed">
                      This price increase carries moderate risk with strong revenue potential. The base scenario suggests a balanced tradeoff between growth and retention. Consider communicating additional value to mitigate customer sensitivity...
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        </div>

        {/* CTA */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.4 }}
          viewport={{ once: true }}
          className="text-center mt-12 md:mt-16"
        >
          <a
            href="https://app.revalyze.com/register"
            className="inline-flex items-center gap-2 px-8 py-4 bg-gradient-to-r from-emerald-500 to-blue-500 text-white font-semibold rounded-xl hover:from-emerald-600 hover:to-blue-600 transition-all shadow-lg shadow-emerald-500/25 hover:shadow-xl hover:shadow-emerald-500/30"
          >
            Try Pricing Simulation Free
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
            </svg>
          </a>
          <p className="text-sm text-slate-500 mt-4">No credit card required • 14-day free trial</p>
        </motion.div>
      </div>
    </section>
  );
};

export default SimulationShowcase;

