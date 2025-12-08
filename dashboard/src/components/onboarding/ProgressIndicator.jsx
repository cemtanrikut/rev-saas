const ProgressIndicator = ({ currentStep, totalSteps, stepLabels }) => {
  // Build steps array from labels
  const steps = stepLabels
    ? stepLabels.map((label, index) => ({ number: index + 1, label }))
    : Array.from({ length: totalSteps }, (_, i) => ({ number: i + 1, label: `Step ${i + 1}` }));

  return (
    <div className="mb-12">
      {/* Step counter */}
      <div className="text-center mb-6">
        <p className="text-sm text-slate-400">
          Step {currentStep} of {totalSteps}
        </p>
      </div>

      {/* Progress bar */}
      <div className="flex items-center justify-between max-w-2xl mx-auto">
        {steps.map((step, index) => (
          <div key={step.number} className="flex items-center flex-1">
            {/* Step circle */}
            <div className="flex flex-col items-center">
              <div
                className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold transition-all ${
                  step.number === currentStep
                    ? 'bg-gradient-to-r from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/50'
                    : step.number < currentStep
                    ? 'bg-blue-500/20 text-blue-400 border border-blue-500/50'
                    : 'bg-slate-700 text-slate-400 border border-slate-600'
                }`}
              >
                {step.number < currentStep ? (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                ) : (
                  step.number
                )}
              </div>
              <span
                className={`mt-2 text-xs font-medium ${
                  step.number === currentStep
                    ? 'text-white'
                    : step.number < currentStep
                    ? 'text-blue-400'
                    : 'text-slate-500'
                }`}
              >
                {step.label}
              </span>
            </div>

            {/* Connector line */}
            {index < steps.length - 1 && (
              <div
                className={`flex-1 h-0.5 mx-2 mt-[-24px] transition-all ${
                  step.number < currentStep
                    ? 'bg-blue-500/50'
                    : 'bg-slate-700'
                }`}
              />
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default ProgressIndicator;
