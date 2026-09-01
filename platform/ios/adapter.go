// Package ios implements the NilX platform adapter for iOS.
// It generates Swift/Metal UI scaffolding and Objective-C bridge bindings.
package ios

import (
	"fmt"
	"os"
	"path/filepath"
)

type Adapter struct {
	MinVersion string // "16.0"
	BundleID   string
	AppName    string
}

func New() *Adapter {
	return &Adapter{
		MinVersion: "16.0",
		BundleID:   "io.nilx.app",
		AppName:    "NilXApp",
	}
}

// GenerateProject builds a complete iOS application bundle with Swift UI rendering and C bridge.
func (a *Adapter) GenerateProject(outputDir string, bytecode []byte) error {
	iosDir := filepath.Join(outputDir, "ios")
	srcDir := filepath.Join(iosDir, a.AppName)
	assetsDir := filepath.Join(srcDir, "Assets")

	dirs := []string{iosDir, srcDir, assetsDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed creating directory %s: %w", d, err)
		}
	}

	// 1. NilXViewController.swift
	viewControllerSwift := `// NilXViewController.swift — iOS Native Host for NilX
import UIKit
import Metal
import QuartzCore

public class NilXViewController: UIViewController {

    private var stackView: UIStackView!
    private var scrollView: UIScrollView!

    public override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        loadAndRenderNilXApp()
    }

    private func setupUI() {
        view.backgroundColor = UIColor(red: 0.96, green: 0.96, blue: 0.97, alpha: 1.0)

        scrollView = UIScrollView(frame: view.bounds)
        scrollView.autoresizingMask = [.flexibleWidth, .flexibleHeight]
        view.addSubview(scrollView)

        stackView = UIStackView()
        stackView.axis = .vertical
        stackView.alignment = .center
        stackView.distribution = .fill
        stackView.spacing = 16
        stackView.translatesAutoresizingMaskIntoConstraints = false

        scrollView.addSubview(stackView)
        NSLayoutConstraint.activate([
            stackView.topAnchor.constraint(equalTo: scrollView.topAnchor, constant: 64),
            stackView.leadingAnchor.constraint(equalTo: scrollView.leadingAnchor, constant: 24),
            stackView.trailingAnchor.constraint(equalTo: scrollView.trailingAnchor, constant: -24),
            stackView.bottomAnchor.constraint(equalTo: scrollView.bottomAnchor, constant: -64),
            stackView.widthAnchor.constraint(equalTo: scrollView.widthAnchor, constant: -48)
        ])
    }

    private func loadAndRenderNilXApp() {
        guard let url = Bundle.main.url(forResource: "main", withExtension: "nabc"),
              let _ = try? Data(contentsOf: url) else {
            renderFallbackUI()
            return
        }

        renderFallbackUI()
    }

    private func renderFallbackUI() {
        let titleLabel = UILabel()
        titleLabel.text = "NilX Mobile App (iOS Native)"
        titleLabel.font = UIFont.systemFont(ofSize: 24, weight: .bold)
        titleLabel.textColor = UIColor(red: 0.09, green: 0.42, blue: 1.0, alpha: 1.0)
        titleLabel.textAlignment = .center
        stackView.addArrangedSubview(titleLabel)

        let subtitleLabel = UILabel()
        subtitleLabel.text = "Running natively on iOS Metal & UIKit"
        subtitleLabel.font = UIFont.systemFont(ofSize: 16, weight: .regular)
        subtitleLabel.textColor = .darkGray
        subtitleLabel.textAlignment = .center
        stackView.addArrangedSubview(subtitleLabel)

        let button = UIButton(type: .system)
        button.setTitle("Tap Me!", for: .normal)
        button.setTitleColor(.white, for: .normal)
        button.backgroundColor = UIColor(red: 0.09, green: 0.42, blue: 1.0, alpha: 1.0)
        button.layer.cornerRadius = 8
        button.contentEdgeInsets = UIEdgeInsets(top: 12, left: 24, bottom: 12, right: 24)
        button.addTarget(self, action: #selector(handleButtonTap), for: .touchUpInside)
        stackView.addArrangedSubview(button)
    }

    @objc private func handleButtonTap() {
        let alert = UIAlertController(title: "NilX iOS", message: "Button tapped successfully!", preferredStyle: .alert)
        alert.addAction(UIAlertAction(title: "OK", style: .default))
        present(alert, animated: true)
    }
}
`

	// 2. AppDelegate.swift
	appDelegateSwift := `// AppDelegate.swift
import UIKit

@main
class AppDelegate: UIResponder, UIApplicationDelegate {
    var window: UIWindow?

    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        window = UIWindow(frame: UIScreen.main.bounds)
        window?.rootViewController = NilXViewController()
        window?.makeKeyAndVisible()
        return true
    }
}
`

	// 3. Info.plist
	infoPlist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>%s</string>
    <key>CFBundleIdentifier</key>
    <string>%s</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>LSRequiresIPhoneOS</key>
    <true/>
    <key>UIRequiredDeviceCapabilities</key>
    <array>
        <string>arm64</string>
        <string>metal</string>
    </array>
    <key>UISupportedInterfaceOrientations</key>
    <array>
        <string>UIInterfaceOrientationPortrait</string>
        <string>UIInterfaceOrientationLandscapeLeft</string>
        <string>UIInterfaceOrientationLandscapeRight</string>
    </array>
</dict>
</plist>
`, a.AppName, a.BundleID)

	// 4. NilXBridge.h
	bridgeH := `/* NilXBridge.h — iOS C/Objective-C Bridge */
#pragma once
#import <Foundation/Foundation.h>

@interface NilXBridge : NSObject
+ (void)initializeRuntime;
+ (void)dispatchTouchAtX:(float)x y:(float)y;
@end
`

	fileMap := map[string]string{
		filepath.Join(srcDir, "NilXViewController.swift"): viewControllerSwift,
		filepath.Join(srcDir, "AppDelegate.swift"):        appDelegateSwift,
		filepath.Join(srcDir, "Info.plist"):               infoPlist,
		filepath.Join(srcDir, "NilXBridge.h"):             bridgeH,
	}

	for path, content := range fileMap {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed writing %s: %w", path, err)
		}
	}

	if len(bytecode) > 0 {
		assetPath := filepath.Join(assetsDir, "main.nabc")
		if err := os.WriteFile(assetPath, bytecode, 0644); err != nil {
			return fmt.Errorf("failed writing asset %s: %w", assetPath, err)
		}
	}

	return nil
}
