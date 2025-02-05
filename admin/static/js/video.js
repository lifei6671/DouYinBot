const wrapper = document.querySelector('.video-wrapper');

// 全局监听body下的play事件
wrapper.addEventListener('play', (e) => {
    const video = e.target;
    if (video.tagName === 'VIDEO') {
        video.volume = 0.1; // 播放时设置音量
    }
}, true); // 使用捕获阶段确保事件触发
wrapper.addEventListener("click", (e) => {
    const video = e.target;
    if (video.tagName === 'VIDEO') {
        if (video.paused) {
            const playPromise = video.play();
            console.log("播放结果", playPromise);
            e.preventDefault();
        } else {
            video.pause();
            e.preventDefault();
        }
    }
}, true)


let videoItems = [];
let currentIndex = 0;
let currentPage = 1;
const pageSize = 2;
let isLoading = false;
let hasMore = true;

let startY = 0;
let currentPosition = 0;
let isAnimating = false;
let isDragging = false;
let wheelDelta = 0;
let lastWheelTime = 0;

// 新增视口高度计算函数
function calculateViewport() {
    const vh = window.innerHeight * 0.01;
    document.documentElement.style.setProperty('--vh', `${vh}px`);

    // 更新所有视频项高度
    document.querySelectorAll('.video-item').forEach(item => {
        item.style.height = `${window.innerHeight}px`;
    });

    // 更新容器高度
    const wrapper = document.querySelector('.video-wrapper');
    wrapper.style.height = `${document.querySelectorAll('.video-item').length * window.innerHeight}px`;

    // 修正当前位置
    wrapper.style.transform = `translateY(-${currentIndex * window.innerHeight}px)`;
}

// 初始化时计算
calculateViewport();

// 优化resize处理
let resizeTimer;
window.addEventListener('resize', () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => {
        calculateViewport();
        wrapper.style.transform = `translateY(-${currentIndex * window.innerHeight}px)`;
    }, 100);
});

// 初始化位置
updatePosition(currentIndex);

// 触摸事件处理
document.addEventListener('touchstart', handleTouchStart);
document.addEventListener('touchmove', handleTouchMove);
document.addEventListener('touchend', handleTouchEnd);

// 鼠标滚轮事件处理
document.addEventListener('wheel', handleWheel, {passive: false});

// 窗口大小变化处理
window.addEventListener('resize', handleResize);

function handleTouchStart(e) {
    if (isAnimating) return;
    startY = e.touches[0].clientY;
    currentPosition = -currentIndex * window.innerHeight;
    wrapper.classList.add('swipe-active');
    isDragging = true;
}

function handleTouchMove(e) {
    // if (!isDragging) return;
    const deltaY = e.touches[0].clientY - startY; // 修正滑动方向
    // updateWrapperPosition(currentPosition + deltaY);
    // 添加边界弹性效果
    // const deltaY = startY - e.touches[0].clientY;
    const newPosition = currentPosition + deltaY;
    const maxPosition = 0;
    const minPosition = -(videoItems.length - 1) * window.innerHeight;
    let clampedPosition = newPosition;

    if (newPosition > maxPosition) {
        clampedPosition = maxPosition + (newPosition - maxPosition) * 0.3;
    } else if (newPosition < minPosition) {
        clampedPosition = minPosition + (newPosition - minPosition) * 0.3;
    }

    wrapper.style.transform = `translateY(${clampedPosition}px)`;
}

function handleTouchEnd(e) {
    if (!isDragging) return;
    isDragging = false;
    wrapper.classList.remove('swipe-active');

    const endY = e.changedTouches[0].clientY;
    const deltaY = startY - endY; // 修正方向计算
    if (deltaY === 0) {
        //如果滚动为0，则用户可能是点击了
        const video = videoItems[currentIndex].querySelector('video');
        if (video.paused) {
            const playPromise = video.play();
            console.log("播放结果", playPromise);
        } else {
            video.pause();
        }
        e.preventDefault();
    } else {
        handleSwipe(deltaY);
    }
}

function handleWheel(e) {
    if (isAnimating) return;

    const now = Date.now();
    const timeDiff = now - lastWheelTime;

    // 限制处理频率（最少50ms处理一次）
    if (timeDiff < 50) return;

    lastWheelTime = now;
    wheelDelta += e.deltaY;

    const threshold = 200;

    console.debug("鼠标滚动了", wheelDelta)
    if (Math.abs(wheelDelta) > threshold) {
        const direction = Math.sign(wheelDelta);
        wheelDelta = 0;
        handleSwipe(direction * threshold);
    }

    e.preventDefault();
}

// 初始化加载第一页
loadMoreVideos();

// 动态创建视频元素
function createVideoElement(videoData) {
    videoID = videoData.video_id;
    const item = document.createElement('div');
    item.className = 'video-item';
    item.innerHTML = `
                <video poster="${videoData.cover}" loop playsinline controls preload="auto">
                    <source src="${videoData.play_addr}" type="video/mp4">
                    <source src="${videoData.local_play_addr}" type="video/mp4">
                     您的浏览器不支持视频播放。
                </video>
                <div class="video-info">
                    <h2 class="video-title"> <a href="${videoData.author_url}" title="${videoData.nickname}">@${videoData.nickname}</a></h2>
                    <p class="video-description">${videoData.desc}</p>
                </div>
            `;
    return item;
}

// 加载更多视频
async function loadMoreVideos() {
    if (isLoading || !hasMore) return;

    showLoading();
    isLoading = true;

    try {
        const params = new URLSearchParams({
            video_id: videoID,
            action: "prev"
        });
        // 实际API调用示例：
        // const response = await fetch(`/api/videos?page=${currentPage}&pageSize=${pageSize}`);
        // const mockData = await response.json();
        let response = await fetch(`${nexURL}?${params.toString()}`);
        if (!response.ok) {
            throw new Error("请求失败");
        }
        const mockData = await response.json();
        if (mockData.errcode === 404) {
            hasMore = false;
            return;
        }
        wrapper.appendChild(createVideoElement(mockData.data));


        videoItems = document.querySelectorAll('.video-item');
        currentPage++;

        // 更新容器高度
        wrapper.style.height = `${videoItems.length * 100}vh`;

    } catch (error) {
        showError();
        console.error('加载失败:', error);
    } finally {
        hideLoading();
        isLoading = false;
    }
}

// 滑动处理函数（修改后）
function handleSwipe(deltaY) {
    const threshold = 150;
    let targetIndex = currentIndex;

    if (Math.abs(deltaY) > threshold) {
        targetIndex = deltaY > 0 ? currentIndex + 1 : currentIndex - 1;
    }
    console.debug("鼠标滑动了", deltaY, targetIndex, currentIndex, videoItems.length);
    // 限制索引范围
    targetIndex = Math.max(0, Math.min(targetIndex, videoItems.length - 1));


    if (targetIndex !== currentIndex) {
        videoItems[currentIndex].querySelector('video').pause();
        currentIndex = targetIndex;
        isAnimating = true;
        animateTransition();

        // 预加载检测
        if (currentIndex >= videoItems.length - 2 && hasMore) {
            loadMoreVideos();
        }
    } else {
        animateRebound();
    }
}

// 显示加载提示
function showLoading() {
    document.getElementById('loadingIndicator').style.display = 'block';
}

// 隐藏加载提示
function hideLoading() {
    document.getElementById('loadingIndicator').style.display = 'none';
}

// 显示错误提示
function showError() {
    document.getElementById('errorIndicator').style.display = 'block';
}

// 重试加载
function retryLoading() {
    document.getElementById('errorIndicator').style.display = 'none';
    loadMoreVideos();
}

function updateWrapperPosition(yPos) {
    const maxPosition = 0;
    const minPosition = -(videoItems.length - 1) * window.innerHeight;

    const clampedPosition = Math.max(minPosition, Math.min(maxPosition, yPos));
    console.debug("用户滑动了: yPost=", yPos, " minPosition=", minPosition, " clampedPosition=", clampedPosition);
    wrapper.style.transform = `translateY(${clampedPosition}px)`;
}

function animateTransition() {
    const targetY = currentIndex * window.innerHeight;
    // wrapper.style.transition = 'transform 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94)';
    // wrapper.style.transform = `translateY(-${targetY}px)`;

    wrapper.addEventListener('transitionend', () => {
        isAnimating = false;
        videoItems[currentIndex].querySelector('video').play();
    }, {once: true});

    wrapper.style.transition = 'transform 0.4s cubic-bezier(0.22, 0.61, 0.36, 1)';
    wrapper.style.transform = `translateY(-${targetY}px)`;

    // 强制重绘修复iOS动画问题
    void wrapper.offsetHeight;
}

function animateRebound() {
    wrapper.style.transition = 'transform 0.3s ease-out';
    wrapper.style.transform = `translateY(-${currentIndex * window.innerHeight}px)`;

    wrapper.addEventListener('transitionend', () => {
        isAnimating = false;
    }, {once: true});
}

function handleResize() {
    updatePosition(currentIndex);
}

function updatePosition(index) {
    currentIndex = index;
    wrapper.style.transform = `translateY(-${index * window.innerHeight}px)`;
}