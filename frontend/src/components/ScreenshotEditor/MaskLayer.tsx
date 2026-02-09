import React, { useRef, useEffect } from 'react';
import './MaskLayer.css';

interface MaskLayerProps {
  imageSrc: string;
  selection: { x: number; y: number; width: number; height: number };
  canvasRef: React.RefObject<HTMLCanvasElement>;
}

/**
 * 遮罩层 - 负责渲染截图和半透明遮罩
 * 职责：
 * 1. 加载并渲染原始截图
 * 2. 绘制半透明遮罩（rgba(0, 0, 0, 0.3)）
 * 3. 清除选中区域的遮罩，保持该区域清晰可见
 *
 * 修复版本：
 * - 修复了依赖数组问题
 * - 正确处理了 selection 变化
 */
export const MaskLayer: React.FC<MaskLayerProps> = ({
  imageSrc,
  selection,
  canvasRef
}) => {
  const imageRef = useRef<HTMLImageElement>(null);

  // 渲染遮罩
  const renderMask = () => {
    const canvas = canvasRef.current;
    const image = imageRef.current;
    if (!canvas || !image) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // 1. 绘制原始图片
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.drawImage(image, 0, 0);

    // 2. 绘制半透明遮罩
    ctx.fillStyle = 'rgba(0, 0, 0, 0.3)';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // 3. 清除选中区域的遮罩
    if (selection.width > 0 && selection.height > 0) {
      ctx.clearRect(selection.x, selection.y, selection.width, selection.height);
      // 重新绘制选中区域的原始图片（保持清晰）
      ctx.drawImage(
        image,
        selection.x, selection.y, selection.width, selection.height,
        selection.x, selection.y, selection.width, selection.height
      );
    }
  };

  // 图片加载和初始化
  useEffect(() => {
    const canvas = canvasRef.current;
    const image = imageRef.current;
    if (!canvas || !image) return;

    // 图片加载完成后渲染
    const handleImageLoad = () => {
      renderMask();
    };

    image.addEventListener('load', handleImageLoad);

    // 如果图片已经加载，直接渲染
    if (image.complete) {
      handleImageLoad();
    }

    return () => {
      image.removeEventListener('load', handleImageLoad);
    };
  }, [imageSrc, canvasRef]);

  // 当选择区域变化时重新渲染
  useEffect(() => {
    renderMask();
  }, [selection]);

  return (
    <>
      <img
        ref={imageRef}
        src={imageSrc}
        alt="Screenshot"
        style={{ display: 'none' }}
      />
      <canvas ref={canvasRef} className="mask-canvas" />
    </>
  );
};

export default MaskLayer;
