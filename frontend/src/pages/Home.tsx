import { Icon } from '../components/Icon'

/**
 * 首页组件
 */
function Home() {
  return (
    <div className="p-8">
      <div className="max-w-2xl mx-auto">
        {/* Hero 区域 */}
        <div className="text-center mb-12 animate-fade-in">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
            <Icon name="sparkles" size={36} color="white" />
          </div>
          <h1 className="text-4xl font-bold mb-4 bg-gradient-to-r from-[#A78BFA] to-[#C4B5FD] bg-clip-text text-transparent">
            LTools
          </h1>
          <p className="text-lg text-white/60 mb-2">
            插件式工具箱
          </p>
          <p className="text-sm text-white/40">
            基于 Wails v3 和 React 的跨平台桌面应用
          </p>
        </div>

        {/* 技术栈信息 */}
        <div className="mt-12 text-center text-white/30 text-sm">
          <p>使用 Wails v3 + React + TypeScript 构建</p>
        </div>
      </div>
    </div>
  )
}

export default Home
