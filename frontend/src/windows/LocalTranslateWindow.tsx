import { JSX } from 'react';
import { LocalTranslateWidget } from '../components/LocalTranslateWidget';

/**
 * LocalTranslateWindow - 独立翻译窗口
 * 用于全局快捷键快速打开
 */
export default function LocalTranslateWindow(): JSX.Element {
  return (
    <div className="h-screen w-screen bg-[#0D0F1A] flex items-center justify-center p-8">
      <div className="w-full max-w-2xl">
        <LocalTranslateWidget />
      </div>
    </div>
  );
}
