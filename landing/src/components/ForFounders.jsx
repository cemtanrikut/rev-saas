import { motion } from 'framer-motion';
import { useInView } from 'framer-motion';
import { useRef } from 'react';

const StageCard = ({ stage, mrrRange, challenges, delay, gradient }) => {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: "-100px" });

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 20 }}
      animate={isInView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      transition={{ duration: 0.5, delay }}
      className="group bg-slate-800/50 backdrop-blur-sm rounded-3xl p-8 border border-slate-700 hover:border-blue-500/50 hover:shadow-xl hover:shadow-blue-500/10 transition-all duration-300"
    >
      <div className="mb-6">
        <h3 className="text-xl font-semibold text-white mb-3 group-hover:text-blue-400 transition-colors">{stage}</h3>
        <div className={`inline-block px-4 py-2 ${gradient} text-white rounded-full text-sm font-semibold shadow-lg`}>
          {mrrRange}
        </div>
      </div>
      <ul className="space-y-4">
        {challenges.map((challenge, index) => (
          <li key={index} className="flex items-start gap-3">
            <div className={`flex-shrink-0 w-6 h-6 ${gradient} rounded-full flex items-center justify-center mt-0.5 shadow-sm`}>
              <svg className="w-3.5 h-3.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={3}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <span className="text-slate-300 leading-relaxed">{challenge}</span>
          </li>
        ))}
      </ul>
    </motion.div>
  );
};

const ForFounders = () => {
  const stages = [
    {
      stage: "Early-stage SaaS",
      mrrRange: "0–10k MRR",
      gradient: "bg-gradient-to-br from-blue-500 to-blue-600",
      challenges: [
        "Should we increase prices or is it too early?",
        "Are we leaving money on the table?",
        "What should our first adjustment look like?"
      ]
    },
    {
      stage: "Growing SaaS",
      mrrRange: "10k–100k MRR",
      gradient: "bg-gradient-to-br from-indigo-500 to-indigo-600",
      challenges: [
        "Are we priced right vs competitors?",
        "Which plan to adjust for maximum impact?",
        "How much MRR can we gain from changes?"
      ]
    },
    {
      stage: "Enterprise SaaS",
      mrrRange: "100k+ MRR",
      gradient: "bg-gradient-to-br from-violet-500 to-violet-600",
      challenges: [
        "How to optimize across segments?",
        "Impact of adding/removing features?",
        "How to increase without hurting retention?"
      ]
    }
  ];

  return (
    <section id="for-founders" className="py-24 md:py-32 bg-slate-800">
      <div className="max-w-7xl mx-auto px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-block px-4 py-1.5 bg-blue-500/10 border border-blue-500/30 rounded-full text-sm font-medium text-blue-400 mb-4">
            For every stage
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            Built for every stage
          </h2>
          <p className="text-xl text-slate-400 max-w-2xl mx-auto">
            From early-stage startups to established enterprises
          </p>
        </motion.div>

        <div className="grid md:grid-cols-3 gap-8 mb-16">
          {stages.map((stage, index) => (
            <StageCard
              key={index}
              stage={stage.stage}
              mrrRange={stage.mrrRange}
              challenges={stage.challenges}
              gradient={stage.gradient}
              delay={index * 0.1}
            />
          ))}
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.3 }}
          viewport={{ once: true }}
          className="text-center max-w-4xl mx-auto"
        >
          <div className="relative">
            <div className="absolute inset-0 bg-gradient-to-r from-blue-500/20 via-indigo-500/20 to-violet-500/20 rounded-3xl blur-3xl"></div>
            <div className="relative bg-gradient-to-r from-slate-800 to-slate-700 border border-slate-700 text-white rounded-3xl p-12 shadow-2xl">
              <h3 className="text-3xl md:text-4xl font-bold mb-4">
                Stop guessing. Start pricing with confidence.
              </h3>
              <p className="text-lg text-slate-300 leading-relaxed">
                Every SaaS company knows pricing is critical. Most still wing it. Revalyze gives you the data you need to price like a pro.
              </p>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default ForFounders;
