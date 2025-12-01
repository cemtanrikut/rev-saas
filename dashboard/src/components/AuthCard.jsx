const AuthCard = ({ children, title, subtitle }) => {
  return (
    <div className="bg-slate-800/50 backdrop-blur-sm rounded-3xl shadow-2xl p-10 border border-slate-700">
      {title && (
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-white mb-2">
            {title}
          </h1>
          {subtitle && (
            <p className="text-slate-400 text-sm">
              {subtitle}
            </p>
          )}
        </div>
      )}
      {children}
    </div>
  );
};

export default AuthCard;

